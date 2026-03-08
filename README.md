# LootSheet

LootSheet is a local-first accounting tool for D&D 5e parties.

The intent is to treat party finances like a small accounting system, but adapted to actual adventuring life instead of modern business bookkeeping. That means double-entry bookkeeping, immutable posted entries, and correction by reversal or adjustment. It also means no invoices, no tax, and plenty of room for accounts like `Wizard Magic Ink` or `Bardic Damages`.

## Status

LootSheet now has a working SQLite-backed CLI foundation. The next major milestone is the TUI shell; packaging, backup/recovery flow, and sample-data polish are still in progress.

Implemented so far:

- local config loading with environment variable overrides
- SQLite initialization from embedded migrations and seed accounts
- database lifecycle detection for `uninitialized`, `current`, `upgradeable`, `foreign`, and `damaged` states
- account create/list/rename/activate/deactivate/delete with posting protection
- journal post/reverse workflows with balancing validation and immutable posted entries
- quest create/list/accept/complete/collect/writeoff lifecycle flows
- loot create/list/appraise/recognize/sell lifecycle flows
- reporting for trial balance, account ledger, quest receivables, promised quests, loot summary, and write-off candidates
- installed-binary-style smoke coverage in `./testapp.sh`
- structured application logging with OTel-backed instrumentation and text levels `DBG`, `INFO`, `WARN`, `ERR`

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
- `src/render` reserved for the upcoming TUI

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
- explicit stdlib CLI parsing for the current command surface
- SQLite for local storage
- `tcell` planned for the full-screen TUI

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
src/render/            upcoming tcell-based TUI shell
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
- `src/app` parses top-level commands and delegates into domain packages
- domain packages own their CLI handlers and data access for their slice
- `src/ledger` owns shared validation, migrations, and DB lifecycle helpers
- `src/config` owns config file parsing, path resolution, and embedded setup assets
- `src/tools` owns shared helpers such as amount parsing and formatting
- `src/render` remains the placeholder for the future TUI shell

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
- `LOOTSHEET_LOG_LEVEL`

## Database Lifecycle

LootSheet stores init-time setup assets in:

- [001_init.sql](/home/raknoel/Projects/Private/Lootsheet/src/config/setup/migrations/001_init.sql)
- [002_add_journal_entry_reversal_tracking.sql](/home/raknoel/Projects/Private/Lootsheet/src/config/setup/migrations/002_add_journal_entry_reversal_tracking.sql)
- [seed_accounts.json](/home/raknoel/Projects/Private/Lootsheet/src/config/setup/seed_accounts.json)

Those files are used only by `lootsheet init` when bootstrapping a fresh SQLite database.
The applied migrations are recorded in SQLite in `schema_migrations`, and `lootsheet db migrate` applies any later embedded migrations to an existing LootSheet database.

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
- `lootsheet journal post|reverse`
- `lootsheet quest create|list|accept|complete|collect|writeoff`
- `lootsheet loot create|list|appraise|recognize|sell`
- `lootsheet report trial-balance|quest-receivables|promised-quests|loot-summary|writeoff-candidates`
- `lootsheet help`

Use `lootsheet help` for the exact command surface and flag syntax.

The repository also includes `./testapp.sh`, which builds a temporary binary and
runs an installed-binary-style smoke scenario against a temporary workspace.

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

## Next Step

The next implementation milestone is the TUI shell around the existing ledger and register workflows.

Near-term supporting work still pending:

- backup-before-migration or repair flows
- sample campaign fixture data
- upgrade and recovery documentation

See [PLAN.md](/home/raknoel/Projects/Private/Lootsheet/PLAN.md), [TODO.md](/home/raknoel/Projects/Private/Lootsheet/TODO.md), and [DESIGN.md](/home/raknoel/Projects/Private/Lootsheet/DESIGN.md) for the working project plan.
