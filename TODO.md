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
- [ ] Replace the external `sqlite3` command dependency with direct SQLite access from the app binary.
- [x] Implement the rule that posted entries cannot be edited or deleted.
- [ ] Expand structured logging configuration beyond the default OTel-backed text logger if needed.

## Next

- [x] Build CLI commands for:
  - [x] database init
  - [x] database status
  - [x] list accounts
  - [x] create account
  - [x] rename account
  - [x] post journal entry
  - [ ] reverse journal entry
  - [ ] create quest
  - [ ] mark quest completed
  - [ ] collect quest payment
  - [ ] create loot appraisal
  - [ ] recognize loot appraisal
- [x] Add account deactivation.
- [ ] Add account ledger output.
- [ ] Add trial balance output.
- [ ] Add loss-on-sale and bad-debt helper flows.

## TUI

- [ ] Set up `tcell` screen initialization and shutdown.
- [ ] Implement alternate-screen lifecycle and terminal resize handling.
- [ ] Define the boxed panel layout system.
- [ ] Build the main dashboard.
- [ ] Build a chart-of-accounts screen.
- [ ] Build a journal entry browser.
- [ ] Build a quest register screen.
- [ ] Build an unrealized loot register screen.
- [ ] Add a theme/color configuration model.
- [ ] Add optional mouse support.
- [ ] Add keyboard navigation and contextual help.

## Data and Rules

- [ ] Ensure account IDs are immutable even when names change.
- [ ] Prevent deletion of accounts that have postings.
- [ ] Allow used accounts to be marked inactive.
- [ ] Keep quest promises off-ledger until earned.
- [ ] Keep unrealized loot appraisals off-ledger until explicitly recognized.
- [ ] Support partial quest payments, advances, and bonuses.
- [ ] Support custom accounts like `Wizard Magic Ink`.

## Reports

- [ ] Trial balance
- [ ] General ledger report
- [ ] Open quest receivables report
- [ ] Promised-but-unearned quest report
- [ ] Unrealized loot summary
- [ ] Write-off candidates

## Quality

- [x] Add a `Makefile` or equivalent development entrypoints.
- [ ] Add `golangci-lint` configuration.
- [ ] Run `go fmt ./...` as a standard check.
- [ ] Run `go vet ./...` as a standard check.
- [ ] Run `golangci-lint run` as a standard check.
- [x] Add unit tests for journal balancing.
- [ ] Add tests for reversal and correction flows.
- [ ] Add tests for quest completion and collection flows.
- [ ] Add tests for loot appraisal recognition and sale flows.
- [ ] Add fixtures with a sample campaign ledger.
- [ ] Add `testapp.sh` end-to-end smoke coverage for installed-binary style runs in a temporary workspace.

## Packaging and Longevity

- [ ] Define supported release targets for Linux amd64/arm64 and macOS amd64/arm64.
- [ ] Decide the first release installation format:
  - [ ] direct binary
  - [ ] archive bundle
  - [ ] optional Homebrew later
- [ ] Finalize the installed file set and locations:
  - [ ] config file
  - [ ] SQLite database
  - [ ] backup directory
  - [ ] optional export directory conventions
- [x] Add full upgrade migration execution beyond init-time migration tracking.
- [ ] Add startup detection for uninitialized, current, upgradeable, foreign, and damaged databases.
- [ ] Add backup creation before risky migration or repair flows.
- [ ] Define stable long-term paths for config, database, backups, and optional exports.
- [ ] Document upgrade and recovery workflow for existing local databases.
- [ ] Evaluate Windows support later after CLI/TUI and packaging are stable on Linux and macOS.

## Later

- [ ] CSV export
- [ ] backup command
- [ ] import sample dataset
- [ ] configurable GP/SP/CP display helpers
- [ ] optional member balance and distribution tracking
