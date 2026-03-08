# LootSheet

LootSheet is a local-first accounting tool for D&D 5e parties.

The intent is to treat party finances like a small accounting system, but adapted to actual adventuring life instead of modern business bookkeeping. That means double-entry bookkeeping, immutable posted entries, and correction by reversal or adjustment. It also means no invoices, no tax, and plenty of room for accounts like `Wizard Magic Ink` or `Bardic Damages`.

## Status

This repository is in the initial bootstrap stage.

Implemented so far:

- Go module initialization
- `main.go -> src/app -> src/config` application bootstrap
- local config loading with environment variable overrides
- core enum definitions for accounts, journal entries, quests, and loot
- embedded setup files for ordered init migrations and default account seeds
- CLI commands for `db status`, `init`, `account list`, and `journal post`
- first-time SQLite initialization that uses setup assets once and then keeps SQLite as the source of truth
- applied init migrations recorded inside SQLite for later inspection
- journal posting with balance validation before persistence
- structured application logging with OTel-backed instrumentation and text levels `DBG`, `INFO`, `WARN`, `ERR`

The first implementation target is:

- Go backend and application logic
- CLI/TUI interface instead of a browser UI
- local SQLite database
- single-user local workflow

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

## Planned Feature Set

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

### Dashboard

- current party cash
- recent posted entries
- open quests
- collectible rewards
- unrealized loot value
- write-off candidates

## Planned Stack

- Go
- Cobra for CLI commands
- `tcell` for low-level terminal rendering, input, mouse, and resize handling
- custom boxed panel renderer inspired by `btop`
- SQLite for local storage

## Planned Repository Shape

```text
main.go                application entrypoint and startup
src/config/            config loading, defaults, env handling
src/app/               dependency wiring and app bootstrap
src/render/            tcell layout, screen model, and drawing
src/service/           core business logic and use cases
src/repo/              SQLite repositories and data access
src/tools/             exports, imports, backup, CSV helpers
docs/                  optional supplementary docs later
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
- `src/config` owns config file parsing and environment handling
- `src/app` wires together config, repositories, services, and renderer
- `src/render` owns terminal rendering and interaction
- `src/service` owns business rules
- `src/repo` owns SQLite access
- `src/tools` holds utility workflows such as CSV export and backup helpers

The expected local checks are:

- `go fmt ./...`
- `go vet ./...`
- `golangci-lint run`

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

## Database Initialization

LootSheet stores init-time setup assets in:

- [001_init.sql](src/config/setup/migrations/001_init.sql)
- [seed_accounts.json](src/config/setup/seed_accounts.json)

Those files are used only by `lootsheet init` when bootstrapping a fresh SQLite database.
The applied init migrations are also recorded in SQLite in `schema_migrations`.

After initialization:

- SQLite is the source of truth for accounts and other stored records
- startup and account listing read from SQLite, not from config seed files
- rerunning `lootsheet init` against an initialized LootSheet database does not reseed it

## CLI

Current commands:

- `lootsheet db status`
- `lootsheet init`
- `lootsheet account list`
- `lootsheet journal post --date YYYY-MM-DD --description TEXT --debit CODE:AMOUNT[:MEMO] --credit CODE:AMOUNT[:MEMO]`

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

The next implementation milestone is to expand the CLI workflows around the new SQLite storage layer:

- account create and rename
- journal entry posting and balancing validation
- quest and loot register commands
- loot appraisals

See [PLAN.md](PLAN.md), [TODO.md](TODO.md), and [DESIGN.md](DESIGN.md) for the working project plan.
