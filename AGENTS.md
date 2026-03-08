# AGENTS.md

This repository is for `LootSheet`, a local-first D&D 5e accounting application.

## Project Intent

Treat the app like accounting software first and a game companion second.

The product should feel credible to someone with bookkeeping experience, while still supporting D&D-specific humor and party-specific edge cases.

## Current Scope

- Go application
- local CLI/TUI
- SQLite storage
- no web frontend
- no auth or network dependency in v1 unless explicitly requested

## Required Reading

Before making significant changes, read:

- [README.md](README.md)
- [DESIGN.md](DESIGN.md)
- [PLAN.md](PLAN.md)
- [TODO.md](TODO.md)

## Hard Invariants

These rules should not be broken unless the user explicitly changes the product direction:

- posted journal entries are immutable
- corrections happen via reversal, adjustment, or reclassification
- journal entries must balance
- account IDs are immutable
- account names are renameable
- accounts with postings cannot be deleted
- used accounts may be marked inactive
- promised quest rewards stay off-ledger until earned
- unrealized loot appraisals stay off-ledger until explicitly recognized

## Domain Guidance

### Ledger vs Registers

Keep the distinction clear:

- the `general ledger` is the formal accounting record
- `registers` track operational state that may later produce ledger entries

Initial registers:

- quest register
- unrealized loot register

Do not collapse these into one generic list unless the user asks for that simplification.

### D&D-Specific Accounting Assumptions

- There is no invoice or VAT model in v1.
- Small consumables such as arrows, ink, and rations should default to expenses.
- High-value loot may be appraised before sale.
- Sales below appraisal should be visible as losses when the appraisal was recognized.

### Custom Accounts

Custom accounts are part of the product, not an extension point for later.

Examples:

- `Wizard Magic Ink`
- `Arrows & Ammunition`
- `Tavern Reparations`

## Engineering Guidance

- Prefer simple, explicit Go over framework-heavy patterns.
- Keep dependencies small and justified.
- Favor SQLite-compatible SQL and storage design.
- Preserve auditability over convenience.
- If adding or changing accounting behavior, update the documentation and tests in the same change.

## Layout Guidance

Use this structure unless the user changes it:

- `main.go` starts the application
- `src/app` for thin CLI routing and bootstrap
- `src/account` for account CRUD, activation, and deletion protection
- `src/journal` for posting, reversal, and account-ledger output
- `src/quest` for quest lifecycle handling
- `src/loot` for loot lifecycle handling
- `src/report` for read-only reporting flows
- `src/ledger` for shared types, validation, DB helpers, errors, and migrations
- `src/config` for config file setup, path resolution, and embedded init assets
- `src/tools` for shared helpers such as currency parsing/formatting
- `src/render` for the upcoming TUI shell

Keep dependency flow one-way:

- `app -> domain packages -> ledger -> config`
- no cross-domain imports between `account`, `journal`, `quest`, `loot`, and `report`

## Preferred Planned Stack

- Go
- explicit stdlib CLI parsing unless there is a strong reason to add a framework
- `tcell`
- SQLite

If you propose alternatives, explain why they are better for this specific repository.

## TUI Guidance

- The visual target is `btop`-inspired boxed panels, not a form-heavy terminal wizard.
- Prefer a cell-based renderer and explicit layout code over a high-level widget stack.
- Theme support, resize handling, and keyboard-first navigation are part of the intended design.
- Mouse support is optional and should never replace keyboard usability.

## Quality Gates

Before considering work complete, prefer to run:

- `make check`

That currently expands to:

- `gofmt -l .`
- `goimports -l .`
- `go test ./...`
- `go vet ./...`
- `golangci-lint run`
- `govulncheck ./...`

## Implementation Priorities

Build in this order unless the user redirects:

1. domain model
2. storage and migrations
3. CLI workflows and reports
4. TUI
5. packaging, backups, and polish

## Editing Expectations

- Keep docs and implementation aligned.
- Do not silently weaken accounting invariants.
- Avoid introducing networked architecture or auth unless explicitly requested.
- When in doubt, prefer a boring design that is easy to test and reason about.
