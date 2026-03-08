# TODO

## Now

- [x] Align the repository to the `main.go + src/...` structure.
- [x] Initialize the Go module.
- [x] Pick the local database file location strategy.
- [x] Implement config file loading in `src/config`.
- [x] Set up application wiring in `src/app`.
- [x] Define the core enum sets:
  - [x] account types
  - [x] journal entry status
  - [x] quest status
  - [x] loot status
- [x] Write the first SQLite schema.
- [x] Seed a default chart of accounts for a D&D party.
- [x] Implement journal balancing validation.
- [x] Replace the external `sqlite3` command dependency with direct SQLite access from the app binary.
- [x] Implement the rule that posted entries cannot be edited or deleted.
- [ ] Expand structured logging configuration beyond the default slog text logger if needed.

## Next

- [x] Build CLI commands for:
  - [x] database init
  - [x] database status
  - [x] list accounts
  - [x] create account
  - [x] rename account
  - [x] post journal entry
  - [x] reverse journal entry
  - [x] create quest
  - [x] mark quest completed
  - [x] collect quest payment
  - [x] create loot appraisal
  - [x] recognize loot appraisal
- [x] Add account deactivation.
- [x] Add account ledger output.
- [x] Add trial balance output.
- [x] Add loss-on-sale and bad-debt helper flows.

## CLI UX

- [x] Add comprehensive hierarchical help across the full CLI tree.
  - [x] `lootsheet help` should explain the main command groups, accounting model, and common workflows.
  - [x] `lootsheet account help`, `lootsheet journal help`, `lootsheet quest help`, `lootsheet loot help`, and `lootsheet report help` should describe each domain workflow clearly.
  - [x] Leaf commands such as `lootsheet account list help` and `lootsheet journal post help` should document flags, required inputs, amount/date formats, and concrete examples.
  - [x] Support consistent `help`, `-h`, and `--help` behavior across top-level, grouped, and leaf commands.
- [x] Generate and ship proper `man` pages for the CLI command tree as part of packaging.

## TUI

- [x] Set up `tcell` screen initialization and shutdown.
- [x] Implement alternate-screen lifecycle and terminal resize handling.
- [x] Define the boxed panel layout system.
- [x] Build the main dashboard.
- [x] Build a chart-of-accounts screen.
- [x] Build a journal entry browser.
- [x] Build a quest register screen.
- [x] Build an unrealized loot register screen.
- [x] Add a theme/color configuration model.
- [x] Add keyboard navigation and contextual help.
- [x] Add shared list selection and detail panes.
- [x] Add the first interactive TUI workflow with account activate/deactivate confirmation.
- [x] Add journal reversal and drill-down workflows in the TUI.
- [x] Add quest collection/write-off workflows in the TUI.
- [x] Add loot recognition workflows in the TUI.
- [x] Add loot sale workflows in the TUI.
- [ ] Add optional mouse support.

## Data and Rules

- [x] Ensure account IDs are immutable even when names change.
- [x] Prevent deletion of accounts that have postings.
- [x] Allow used accounts to be marked inactive.
- [x] Keep quest promises off-ledger until earned.
- [x] Keep unrealized loot appraisals off-ledger until explicitly recognized.
- [x] Support partial quest payments, advances, and bonuses.
- [x] Support custom accounts like `Wizard Magic Ink`.

## Reports

- [x] Trial balance
- [x] General ledger report
- [x] Open quest receivables report
- [x] Promised-but-unearned quest report
- [x] Unrealized loot summary
- [x] Write-off candidates

## Quality

- [x] Add a `Makefile` or equivalent development entrypoints.
- [x] Add `golangci-lint` configuration.
- [x] Run `go fmt ./...` as a standard check.
- [x] Run `go vet ./...` as a standard check.
- [x] Run `golangci-lint run` as a standard check.
- [x] Add unit tests for journal balancing.
- [x] Add tests for reversal and correction flows.
- [x] Add tests for quest completion and collection flows.
- [x] Add tests for loot appraisal recognition and sale flows.
- [x] Add fixtures with a sample campaign ledger.
- [x] Add `testapp.sh` end-to-end smoke coverage for installed-binary style runs in a temporary workspace.

## Packaging and Longevity

- [ ] Define supported release targets for Linux amd64/arm64 and macOS amd64/arm64.
- [ ] Decide the first release installation format:
  - [ ] direct binary
  - [ ] archive bundle
  - [ ] optional Homebrew later
- [ ] Finalize the installed file set and locations:
  - [ ] config file
  - [ ] SQLite database
  - [x] backup directory
  - [x] optional export directory conventions
- [x] Add full upgrade migration execution beyond init-time migration tracking.
- [x] Add startup detection for uninitialized, current, upgradeable, foreign, and damaged databases.
- [x] Add backup creation before risky migration or repair flows.
- [x] Define stable long-term paths for config, database, backups, and optional exports.
- [x] Document upgrade and recovery workflow for existing local databases.
- [ ] Evaluate Windows support later after CLI/TUI and packaging are stable on Linux and macOS.

## Later

- [ ] CSV export
- [ ] backup command
- [ ] import sample dataset
- [ ] configurable GP/SP/CP display helpers
- [ ] optional member balance and distribution tracking
