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
- [ ] Implement journal balancing validation.
- [ ] Implement the rule that posted entries cannot be edited or deleted.
- [ ] Expand structured logging configuration beyond the default OTel-backed text logger if needed.

## Next

- [x] Build CLI commands for:
  - [x] database init
  - [x] list accounts
  - [ ] create account
  - [ ] rename account
  - [ ] post journal entry
  - [ ] reverse journal entry
  - [ ] create quest
  - [ ] mark quest completed
  - [ ] collect quest payment
  - [ ] create loot appraisal
  - [ ] recognize loot appraisal
- [ ] Add account deactivation.
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

- [ ] Add a `Makefile` or equivalent development entrypoints.
- [ ] Add `golangci-lint` configuration.
- [ ] Run `go fmt ./...` as a standard check.
- [ ] Run `go vet ./...` as a standard check.
- [ ] Run `golangci-lint run` as a standard check.
- [ ] Add unit tests for journal balancing.
- [ ] Add tests for reversal and correction flows.
- [ ] Add tests for quest completion and collection flows.
- [ ] Add tests for loot appraisal recognition and sale flows.
- [ ] Add fixtures with a sample campaign ledger.

## Later

- [ ] CSV export
- [ ] backup command
- [ ] import sample dataset
- [ ] configurable GP/SP/CP display helpers
- [ ] optional member balance and distribution tracking
