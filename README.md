# LootSheet

LootSheet is a local-first accounting tool for D&D 5e parties.

The intent is to treat party finances like a small accounting system, but adapted to actual adventuring life instead of modern business bookkeeping. That means double-entry bookkeeping, immutable posted entries, and correction by reversal or adjustment. It also means no invoices, no tax, and plenty of room for accounts like `Wizard Magic Ink` or `Bardic Damages`.

## Status

LootSheet now has a working SQLite-backed CLI foundation plus an interactive multi-screen TUI slice. The TUI opens into a boxed dashboard, moves between Accounts, Journal, Quests, and Loot screens with keyboard navigation, keeps a selected row and detail pane on list screens, supports account activate/deactivate, journal reversal, full-balance quest collection/write-off, latest-appraisal loot recognition, recognized-loot sale with amount entry, and guided expense/income/custom journal entry creation from the dashboard, and redraws cleanly on resize while staying backed by app-facing adapters; packaging, backup/recovery flow, and sample-data polish are still in progress.

Implemented so far:

- local config loading with environment variable overrides
- SQLite initialization from embedded migrations and seed accounts
- database lifecycle detection for `uninitialized`, `current`, `upgradeable`, `foreign`, and `damaged` states
- account create/list/rename/activate/deactivate/delete with posting protection
- guided `entry expense|income|custom` workflows for common journal creation
- journal post/reverse workflows with balancing validation and immutable posted entries
- quest create/list/accept/complete/collect/writeoff lifecycle flows
- loot create/list/appraise/recognize/sell lifecycle flows
- reporting for trial balance, account ledger, quest receivables, promised quests, loot summary, and write-off candidates
- interactive `tcell`-backed TUI shell with alternate-screen lifecycle, resize-aware boxed panels, contextual footer help, list selection/detail panes, an Accounts activate/deactivate workflow, a Journal line-detail/reversal workflow, Quest full-balance collect/write-off actions, and Loot recognition/sale actions backed by app/domain read models
- installed-binary-style smoke coverage in `./testapp.sh`
- structured application logging via stdlib `slog` with text levels `DBG`, `INFO`, `WARN`, `ERR`

## Current Architecture

The codebase is organized as vertical slices under `src/`:

- `src/app` for bootstrap and top-level CLI routing
- `src/account` for account CRUD and account state changes
- `src/journal` for posting, reversal, and ledger output
- `src/quest` for quest lifecycle tracking
- `src/loot` for loot lifecycle tracking
- `src/report` for read-only reports
- `src/ledger` for shared domain types, validation, DB helpers, errors, and migrations
- `src/config` for config loading and embedded setup assets
- `src/tools` for shared helpers such as D&D currency parsing/formatting
- `src/render` for the `tcell`-backed TUI shell and rendering primitives

Dependency flow is intentionally one-way:

- `app -> domain packages -> ledger -> config`
- no cross-domain imports between domain slices

All amounts are stored as `int64` copper pieces and exposed through `tools.ParseAmount` and `tools.FormatAmount`.

## Product Goals

- Keep party finances auditable and funny at the same time.
- Make the app feel like real accounting software, not just a loot tracker.
- Separate operational lists from the formal ledger when appropriate.
- Support custom accounts and party-specific bookkeeping habits.
- Stay local, simple, and maintainable.

## Core Principles

### 1. Posted journal entries are immutable

Once an entry is posted, it is never edited or deleted.

Mistakes are handled by:

- reversal entries
- adjusting entries
- reclassification entries

### 2. Not everything belongs in the ledger immediately

LootSheet distinguishes between:

- the `general ledger`
- operational `registers`

Examples:

- A promised quest reward belongs in the `quest register` until it is actually earned.
- An appraised magic item can live in the `unrealized loot register` until the party decides to recognize it in the ledger.

### 3. D&D economics are not normal business economics

The app assumes:

- no VAT or tax logic
- most quest promises are conditional
- loot values are uncertain until sold
- small consumables are expenses, not durable assets

### 4. User-defined accounts are a feature, not an edge case

The app must support custom accounts such as:

- `Wizard Magic Ink`
- `Arrows & Ammunition`
- `Tavern Reparations`
- `Bardic Legal Exposure`

Accounts may be renamed. Their internal IDs must remain stable.

These are bookkeeping accounts in the chart of accounts, not login identities.
LootSheet does not have a user-account or auth model in v1.

## Feature Set

### Ledger

- chart of accounts
- journal entry posting
- account ledgers
- trial balance
- reversal and correction workflow

### Registers

- quest register
- unrealized loot register
- optionally party supplies or distribution register later

### Reports

- trial balance
- account ledger
- outstanding quest receivables
- promised-but-unearned quests
- unrealized loot summary
- write-off candidates

## Runtime Stack

- Go
- Cobra for the CLI command tree and help routing
- SQLite for local storage
- `tcell` for the full-screen TUI shell

## Repository Shape

```text
main.go                application entrypoint and startup
src/config/            config loading, defaults, env handling
src/app/               dependency wiring and app bootstrap
src/account/           account commands and persistence
src/journal/           journal commands and ledger reporting
src/quest/             quest register commands and persistence
src/loot/              loot register commands and persistence
src/report/            reporting commands and queries
src/ledger/            shared types, validation, DB helpers, migrations
src/render/            tcell-based TUI shell and rendering primitives
src/tools/             currency and utility helpers
```

## TUI Direction

The TUI target is `btop`-inspired rather than form-based.

That means:

- full-screen terminal UI
- boxed panels and dense dashboard layout
- theme-driven colors
- Unicode box drawing and graph glyphs where useful
- keyboard-first navigation with optional mouse support

This does not require C++. The visual style comes from terminal rendering strategy rather than language choice.

## Development Workflow

Development should stay boring and explicit.

- `main.go` starts the application and hands off to `src/app`
- `src/app` owns the Cobra command tree, leaf flag parsing, and top-level help
- domain packages own command execution, validation, output rendering, and data access for their slice
- `src/ledger` owns shared validation, migrations, and DB lifecycle helpers
- `src/config` owns config file parsing, path resolution, and embedded setup assets
- `src/tools` owns shared helpers such as amount parsing and formatting
- `src/render` owns the cell renderer, layout primitives, and interactive TUI shell

The preferred local checks are:

- `make check`
- `bash ./testapp.sh`

## Configuration

LootSheet currently loads an optional JSON config file from:

- Linux and other XDG-style systems: `$XDG_CONFIG_HOME/lootsheet/config.json` or `~/.config/lootsheet/config.json`
- macOS: `~/Library/Application Support/lootsheet/config.json`
- Windows: the OS user config directory for `lootsheet`

The default database path is:

- Linux and other XDG-style systems: `$XDG_DATA_HOME/lootsheet/lootsheet.db` or `~/.local/share/lootsheet/lootsheet.db`
- macOS: `~/Library/Application Support/lootsheet/lootsheet.db`
- Windows: `%LOCALAPPDATA%\\lootsheet\\lootsheet.db` when available

Environment overrides:

- `LOOTSHEET_CONFIG`
- `LOOTSHEET_DATA_DIR`
- `LOOTSHEET_DATABASE_PATH`
- `LOOTSHEET_BACKUP_DIR`
- `LOOTSHEET_EXPORT_DIR`
- `LOOTSHEET_LOG_LEVEL`

Resolved long-term storage paths:

- config file under the OS user config directory
- primary SQLite database under the OS user data directory
- backups under `backups/` inside the LootSheet data directory by default
- exports under `exports/` inside the LootSheet data directory by default

## Database Lifecycle

LootSheet stores init-time setup assets in:

- [001_init.sql](src/config/setup/migrations/001_init.sql)
- [002_add_journal_entry_reversal_tracking.sql](src/config/setup/migrations/002_add_journal_entry_reversal_tracking.sql)
- [seed_accounts.json](src/config/setup/seed_accounts.json)

Those files are used only by `lootsheet init` when bootstrapping a fresh SQLite database.
The applied migrations are recorded in SQLite in `schema_migrations`, and `lootsheet db migrate` applies any later embedded migrations to an existing LootSheet database.

When `lootsheet db migrate` is about to apply a schema migration or repair legacy metadata, it first writes a timestamped backup copy into the configured backup directory.

`lootsheet db status` classifies the current database as:

- `uninitialized`
- `current`
- `upgradeable`
- `foreign`
- `damaged`

After initialization:

- SQLite is the source of truth for accounts and other stored records
- startup and account listing read from SQLite, not from config seed files
- rerunning `lootsheet init` against an initialized LootSheet database does not reseed it
- foreign or damaged databases are reported clearly and are not migrated implicitly

## CLI

Current commands:

- `lootsheet db status`
- `lootsheet db migrate`
- `lootsheet init`
- `lootsheet account list|create|rename|deactivate|activate|delete|ledger`
- `lootsheet entry expense|income|custom`
- `lootsheet journal post|reverse`
- `lootsheet quest create|list|accept|complete|collect|writeoff`
- `lootsheet loot create|list|appraise|recognize|sell`
- `lootsheet report trial-balance|quest-receivables|promised-quests|loot-summary|writeoff-candidates`
- `lootsheet help`

Use `lootsheet help` for the exact command surface and flag syntax.
The CLI command tree is now routed through Cobra while the domain slices still own the underlying command handlers and accounting behavior.

Help is hierarchical across the command tree:

- `lootsheet help`
- `lootsheet account help`
- `lootsheet account list help`
- `lootsheet entry expense -h`
- `lootsheet journal post --help`

The repository also includes `./testapp.sh`, which builds a temporary binary and
runs an installed-binary-style smoke scenario against a temporary workspace.

Generated CLI man pages are checked in under [docs/man/](docs/man) and can be refreshed with:

```sh
make manpages
```

## Accounting Examples

### Sell loot below appraisal

If loot was recognized at 500 GP and later sold for 150 GP:

```text
Dr Party Cash                150
Dr Loss on Sale of Loot      350
Cr Loot Inventory            500
```

### Quest completed, then collected

```text
Dr Quest Receivable          100
Cr Quest Income              100
```

```text
Dr Party Cash                100
Cr Quest Receivable          100
```

### Quest reward becomes uncollectible

```text
Dr Failed Patron Loss        100
Cr Quest Receivable          100
```

## What This Project Is Not

- a general business accounting package
- a multiplayer SaaS product
- a browser-first app
- a full inventory simulator for every torch and rope unless the user wants that level of detail

## Sample Fixture

The repository now includes a reusable sample campaign fixture at [fixtures/sample_campaign.sql](fixtures/sample_campaign.sql).

It is currently used for regression coverage and documentation, and captures:

- a custom account
- open and accepted quest promises
- a stale partially paid receivable
- held, recognized, and sold loot
- a loss-on-sale example

There is not yet a user-facing import command for that fixture; `import sample dataset` remains a later CLI/TUI backlog item.

## Upgrade And Recovery

The current supported database upgrade and recovery workflow is documented in [docs/upgrade-recovery.md](docs/upgrade-recovery.md).

In short:

- use `lootsheet db status` to confirm whether the database is `uninitialized`, `current`, `upgradeable`, `foreign`, or `damaged`
- use `lootsheet db migrate` only for `upgradeable` databases
- rely on the timestamped backup path printed by `db migrate` before any risky schema change
- do not try to migrate `foreign` or `damaged` databases in place

## Next Step

The next implementation milestone is packaging and operational workflows on top of the current CLI/TUI surface.

Near-term supporting work still pending:

- release target and installation decisions
- backup/export workflows
- packaging polish around generated man pages

See [PLAN.md](PLAN.md), [TODO.md](TODO.md), and [DESIGN.md](DESIGN.md) for the working project plan.
