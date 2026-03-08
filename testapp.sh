#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
SMOKE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/lootsheet-smoke.XXXXXX")"
BINARY_PATH="$SMOKE_DIR/lootsheet"

cleanup() {
	if [ "${KEEP_SMOKE_DIR:-0}" = "1" ]; then
		printf 'Keeping smoke workspace at %s\n' "$SMOKE_DIR"
		return
	fi

	rm -rf "$SMOKE_DIR"
}
trap cleanup EXIT

export GOCACHE="${GOCACHE:-$SMOKE_DIR/go-cache}"
export LOOTSHEET_CONFIG="$SMOKE_DIR/config/config.json"
export LOOTSHEET_DATA_DIR="$SMOKE_DIR/data"
export LOOTSHEET_DATABASE_PATH="$SMOKE_DIR/data/lootsheet.db"

LAST_OUTPUT=""

note() {
	printf '==> %s\n' "$*"
}

fail() {
	printf 'smoke test failed: %s\n' "$*" >&2
	exit 1
}

run_ok() {
	local output

	note "$*"
	if ! output="$("$BINARY_PATH" "$@" 2>&1)"; then
		printf '%s\n' "$output" >&2
		fail "command exited non-zero: $*"
	fi

	LAST_OUTPUT="$output"
	printf '%s\n' "$output"
}

run_fail() {
	local output

	note "$*"
	if output="$("$BINARY_PATH" "$@" 2>&1)"; then
		printf '%s\n' "$output" >&2
		fail "command unexpectedly succeeded: $*"
	fi

	LAST_OUTPUT="$output"
	printf '%s\n' "$output"
}

assert_contains() {
	local haystack="$1"
	local needle="$2"
	local context="$3"

	case "$haystack" in
		*"$needle"*) ;;
		*)
			printf 'Output was:\n%s\n' "$haystack" >&2
			fail "$context"
			;;
	esac
}

assert_not_contains() {
	local haystack="$1"
	local needle="$2"
	local context="$3"

	case "$haystack" in
		*"$needle"*)
			printf 'Output was:\n%s\n' "$haystack" >&2
			fail "$context"
			;;
		*) ;;
	esac
}

extract_line_field() {
	local input="$1"
	local pattern="$2"
	local field="$3"

	printf '%s\n' "$input" | awk -v pattern="$pattern" -v field="$field" '$0 ~ pattern {print $field; exit}'
}

require_nonempty() {
	local value="$1"
	local label="$2"

	if [ -z "$value" ]; then
		fail "$label was empty"
	fi
}

note "Building LootSheet binary into temporary workspace"
(cd "$ROOT_DIR" && go build -buildvcs=false -o "$BINARY_PATH" .)

run_ok db status
assert_contains "$LAST_OUTPUT" "State: uninitialized" "db status should report uninitialized before init"

run_ok init
assert_contains "$LAST_OUTPUT" "LootSheet initialized" "init should initialize the database"

run_ok db status
assert_contains "$LAST_OUTPUT" "State: current" "db status should report current after init"

run_ok account create --code 2190 --name "Tavern Reparations Reserve" --type liability
assert_contains "$LAST_OUTPUT" "Created account" "account create should succeed"

run_ok account list
assert_contains "$LAST_OUTPUT" "Tavern Reparations Reserve" "account list should show the custom account"

run_ok account deactivate --code 2190
assert_contains "$LAST_OUTPUT" "Deactivated account 2190" "account deactivate should succeed"

run_ok account activate --code 2190
assert_contains "$LAST_OUTPUT" "Activated account 2190" "account activate should succeed"

run_fail journal post --date 2026-03-01 --description "Broken smoke entry" --debit 5000:2GP --credit 1000:1GP
assert_contains "$LAST_OUTPUT" "journal entry is not balanced" "unbalanced journal post should fail"

run_ok journal post --date 2026-03-01 --description "Restock party supplies" --debit 5000:3GP:Arrows --credit 1000:3GP
assert_contains "$LAST_OUTPUT" "Posted journal entry #1" "balanced journal post should succeed"

run_ok report trial-balance
assert_contains "$LAST_OUTPUT" "Trial Balance" "trial balance report should render"
assert_contains "$LAST_OUTPUT" "BALANCED" "trial balance should remain balanced"

run_ok quest create --title "Bandit Bounty" --patron "Mayor Elra" --reward 7GP --advance 1GP --bonus "Bonus if prisoners are returned alive" --status offered
assert_contains "$LAST_OUTPUT" "Created quest" "offered quest creation should succeed"

run_ok report promised-quests
assert_contains "$LAST_OUTPUT" "Bandit Bounty" "promised quests report should list offered quests"
assert_contains "$LAST_OUTPUT" "7 GP" "promised quests report should show promised reward"

run_ok quest create --title "Old Debt" --patron "Count Ves" --reward 5GP --status accepted --accepted-on 2026-01-01
assert_contains "$LAST_OUTPUT" "Created quest" "accepted quest creation should succeed"
QUEST_ID="$(extract_line_field "$LAST_OUTPUT" '^Created quest ' 3)"
require_nonempty "$QUEST_ID" "quest id"

run_ok quest complete --id "$QUEST_ID" --date 2026-01-02
assert_contains "$LAST_OUTPUT" "Completed quest" "quest completion should succeed"

run_ok quest collect --id "$QUEST_ID" --amount 2GP --date 2026-01-10 --description "Count Ves sent only a partial payment"
assert_contains "$LAST_OUTPUT" "Collected quest payment as journal entry" "quest collection should succeed"

run_ok report quest-receivables
assert_contains "$LAST_OUTPUT" "Old Debt" "quest receivables should list the partially paid quest"
assert_contains "$LAST_OUTPUT" "3 GP" "quest receivables should show the outstanding balance"

run_ok report writeoff-candidates --as-of 2026-03-15 --min-age-days 30
assert_contains "$LAST_OUTPUT" "Old Debt" "write-off candidates should include stale receivables"
assert_contains "$LAST_OUTPUT" "72" "write-off candidates should include receivable age"

run_ok quest writeoff --id "$QUEST_ID" --date 2026-03-20
assert_contains "$LAST_OUTPUT" "Wrote off quest as journal entry" "quest write-off should succeed"

run_ok report quest-receivables
assert_contains "$LAST_OUTPUT" "No outstanding quest receivables." "written-off quest should drop from receivables report"

run_ok loot create --name "Ruby Idol" --source "Old Debt" --quantity 1 --holder "Quartermaster"
assert_contains "$LAST_OUTPUT" "Created loot item" "loot create should succeed"
LOOT_ID="$(extract_line_field "$LAST_OUTPUT" '^Created loot item ' 4)"
require_nonempty "$LOOT_ID" "loot id"

run_ok loot appraise --id "$LOOT_ID" --value 5GP --date 2026-03-21 --appraiser "Guild Assayer"
assert_contains "$LAST_OUTPUT" "Appraised loot item" "loot appraisal should succeed"
APPRAISAL_ID="$(extract_line_field "$LAST_OUTPUT" '^Appraisal ID: ' 3)"
require_nonempty "$APPRAISAL_ID" "appraisal id"

run_ok report loot-summary
assert_contains "$LAST_OUTPUT" "Ruby Idol" "loot summary should list held loot"
assert_contains "$LAST_OUTPUT" "5 GP" "loot summary should show the latest appraisal"

run_ok loot recognize --appraisal-id "$APPRAISAL_ID" --date 2026-03-22
assert_contains "$LAST_OUTPUT" "Recognized loot appraisal as journal entry" "loot recognition should succeed"

run_ok loot sell --id "$LOOT_ID" --amount 3GP --date 2026-03-23 --description "Sold ruby idol under pressure"
assert_contains "$LAST_OUTPUT" "Sold loot item as journal entry" "loot sale should succeed"

run_ok report loot-summary
assert_not_contains "$LAST_OUTPUT" "Ruby Idol" "sold loot should not remain in the held/recognized summary"

run_ok report trial-balance
assert_contains "$LAST_OUTPUT" "BALANCED" "books should remain balanced after the full smoke flow"

printf 'Smoke test passed. Workspace: %s\n' "$SMOKE_DIR"
