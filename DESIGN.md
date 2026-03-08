# Design

## Purpose

LootSheet is a local accounting application for D&D 5e parties. It should feel like actual bookkeeping software while remaining adapted to adventuring realities.

The product is intentionally not a web app in v1. It is a local Go application with a CLI/TUI interface and a local SQLite database.

## Design Goals

- Preserve real accounting behavior where it matters.
- Keep the general ledger strict and auditable.
- Let uncertain or conditional events live in operational registers first.
- Make the app useful during or between sessions.
- Support humor and party-specific quirks through flexible accounts and naming.

## Application Structure

The intended code layout is:

- `main.go` for application startup
- `src/config` for config file loading and environment handling
- `src/app` for dependency injection and runtime bootstrap
- `src/render` for `tcell` rendering, layout, and input coordination
- `src/service` for business logic
- `src/repo` for SQLite repositories and SQL access
- `src/tools` for CSV export, backup, and similar utilities

The dependency direction should stay simple:

- `main.go` depends on `src/app`
- `src/app` depends on `config`, `repo`, `service`, and `render`
- `service` depends on `repo` interfaces or repository layer
- `render` depends on `service` or app-facing interfaces, not raw SQL
- `repo` depends on SQLite and storage concerns only

This is intentionally straightforward DI rather than framework-driven wiring.

## Install and Runtime Model

LootSheet should be designed as a compiled local application for proper desktop/server-style operating systems rather than as a development-only tool.

Initial supported platforms:

- Linux
- macOS

Later if the maintenance burden stays reasonable:

- Windows

The operational model should be:

- a single local binary the user installs and runs directly
- no required background daemon
- no required network service
- no auth or login-user model
- one local SQLite database as the system of record for the party books
- no runtime dependency on an external `sqlite3` executable

### Filesystem Layout

The application should keep long-term files in stable per-user locations:

- config in the user config directory
- SQLite database in the user data directory
- backups in a predictable application-owned backup location under user data or state
- smoke-test or temp working files only in explicit temporary locations
- optional exports in user-chosen locations

These locations should remain stable across upgrades unless the user explicitly changes them.

### Upgrade and Migration Model

Long-term use requires a conservative storage policy:

- every database schema change must be versioned
- startup must detect whether the database is uninitialized, current, upgradeable, foreign, or damaged
- migrations should be forward-only and explicit
- risky migrations should create or require a backup first
- failures should stop with a recovery message rather than partially mutating data silently

### Logging and Diagnostics

The default operational logging should be human-readable and predictable:

- logs to stderr by default
- level names normalized to `DBG`, `INFO`, `WARN`, `ERR`
- structured enough to support machine verification in smoke scripts
- future support for log-file output only if it is justified by real operational needs

### Distribution

The preferred release path for v1 should be simple:

- publish versioned binaries for Linux and macOS
- support archive-based installation first
- add package-manager integrations later only if they reduce real friction

The release artifact should remain self-contained from a storage point of view:

- the `lootsheet` binary should open SQLite directly through an embedded driver
- installed operation should not require users to separately install `sqlite3`
- the app may create config, database, backup, and optional export files, but not depend on helper executables for normal operation

The application should remain usable for years even if package-manager integration changes, which means direct binary installs and stable on-disk data layout matter more than fancy installers.

## Core Model

LootSheet has two main layers:

### 1. General Ledger

The ledger contains formal accounting records:

- accounts
- journal entries
- journal lines

Rules:

- journal entries must balance
- posted entries are immutable
- corrections happen through reversal or adjustment
- account IDs are immutable
- account names are editable
- accounts with postings cannot be deleted
- accounts with postings may be marked inactive

### 2. Operational Registers

Registers track real information that is not always ready for ledger recognition.

Initial registers:

- quest register
- unrealized loot register

These registers are first-class features, not temporary hacks.

## Why Registers Exist

Some D&D facts are useful but not yet bookable:

- a patron promises payment if the party succeeds
- a gem is appraised at 500 GP but has not been sold
- a dragon offers a bonus only if the bard does not insult it again

Those belong in operational tracking first. They enter the ledger only when an accounting event actually occurs.

## Accounting Rules

### Journal Entry Immutability

States:

- `draft`
- `posted`
- `reversed`

Behavior:

- drafts may be edited
- posted entries may not be edited
- reversed entries remain visible
- every correction leaves an audit trail

### Account Types

Initial supported types:

- `asset`
- `liability`
- `equity`
- `income`
- `expense`

Contra accounts can be supported later through configuration or signage rules, but v1 can avoid over-modeling them if needed.

## Suggested Seed Accounts

Assets:

- `Party Cash`
- `Quest Receivable`
- `Loot Inventory`
- `Gear Inventory`

Liabilities:

- `Unearned Quest Revenue`

Equity:

- `Party Equity`

Income:

- `Quest Income`
- `Bonus Quest Income`
- `Unrealized Loot Gain`
- `Gain on Sale of Loot`

Expenses:

- `Adventuring Supplies`
- `Arrows & Ammunition`
- `Wizard Magic Ink`
- `Inn & Travel`
- `Loss on Sale of Loot`
- `Failed Patron Loss`

The seed chart is a starting point only. Users must be able to add custom accounts.

The seed chart should live in init-time setup files rather than long-lived in-memory defaults.
Those setup files are used to populate a fresh SQLite database, and after that the database is the source of truth.

These accounts are ledger accounts. They are not user/login accounts, and v1 should not introduce an auth model.

## Quest Register

The quest register tracks work that may or may not produce income.

Each quest should capture:

- title
- patron
- description
- promised base reward
- optional partial advance
- optional bonus conditions
- current status
- related journal entries
- notes

Suggested statuses:

- `offered`
- `accepted`
- `completed`
- `collectible`
- `partially_paid`
- `paid`
- `defaulted`
- `voided`

### Quest Recognition Rules

Promised but unearned reward:

- stays off-ledger

Advance received before completion:

```text
Dr Party Cash
Cr Unearned Quest Revenue
```

Reward earned on completion:

```text
Dr Quest Receivable
Cr Quest Income
```

If there was an advance:

```text
Dr Unearned Quest Revenue
Cr Quest Income
```

Collection:

```text
Dr Party Cash
Cr Quest Receivable
```

Write-off after recognition:

```text
Dr Failed Patron Loss
Cr Quest Receivable
```

If the quest fails before income is earned, the quest can be closed in the register without any ledger entry.

## Unrealized Loot Register

The unrealized loot register tracks found or appraised items before disposition.

Each loot item should capture:

- item name
- source
- quantity
- appraised value
- appraisal notes
- holder or location
- current status
- related journal entries

Suggested statuses:

- `held`
- `recognized`
- `sold`
- `assigned`
- `consumed`
- `discarded`

### Loot Recognition Rules

Appraisal only:

- stays off-ledger by default

Optional recognition of appraised value:

```text
Dr Loot Inventory
Cr Unrealized Loot Gain
```

Sale below recognized value:

```text
Dr Party Cash
Dr Loss on Sale of Loot
Cr Loot Inventory
```

Sale above recognized value:

```text
Dr Party Cash
Cr Loot Inventory
Cr Gain on Sale of Loot
```

If the party never recognizes the appraisal, then the eventual sale can be recorded directly as realized income or cash movement depending on how strict the workflow should be.

## Consumables and Supplies

By default, low-value consumables should be expensed immediately rather than tracked as assets.

Examples:

- arrows
- rations
- lamp oil
- chalk
- wizard ink

Typical entry:

```text
Dr Adventuring Supplies
Cr Party Cash
```

This should remain flexible. A user may choose to create inventory accounts for more detailed campaigns.

## Custom Accounts

Custom accounts are required for the product to feel alive.

Requirements:

- users can create accounts
- users can rename accounts
- account IDs never change
- accounts with activity cannot be deleted
- accounts can be deactivated
- accounts should support optional parent grouping later

Examples:

- `Wizard Magic Ink`
- `Dragon Courtship Expenses`
- `Clerical Resurrection Fund`

## Dashboard Design

The dashboard should combine formal accounting and operational reminders.

Primary widgets:

- party cash summary
- recent posted entries
- open quest count
- collectible quest total
- unrealized loot appraisal total
- overdue or stale collectible quests
- pending write-off candidates

The dashboard should answer:

- how much cash the party has now
- what is probably collectible
- what is only promised
- what loot has value but is unsold
- what needs correction or write-off

## TUI Design

Initial screens:

- Dashboard
- Accounts
- Journal
- Quests
- Loot
- Reports

Navigation should be keyboard-first and obvious. The design target is closer to `btop` than to a form wizard or roguelike.

Potential layout:

- boxed panels with clear borders and titles
- dense dashboard with multiple simultaneous summaries
- small help footer
- modal or panel flows for entry creation and correction

### Rendering Approach

The TUI should be implemented with `tcell`.

Reasons:

- low-level control over terminal cells
- clean handling of resize, mouse, and keyboard input
- better fit for a `btop`-style boxed layout than a higher-level widget stack
- no need for C++ to achieve the desired look

The intended renderer behavior is:

- use the alternate screen
- maintain a screen model and redraw efficiently
- support Unicode box-drawing characters
- support themeable colors
- degrade gracefully on lower-color terminals

### Visual Direction

The reference style is inspired by `btop`, but LootSheet should still look like accounting software.

That means:

- framed panels with strong section titles
- compact summaries at the top
- dense but readable tables in the main work area
- visual emphasis on balances, exceptions, and collectible items
- restrained animations if any; clarity is more important than spectacle

### Dashboard Layout

The dashboard should feel information-dense on one screen.

Suggested panel layout:

- top-left: `Party Cash`, key balances, and current period summary
- top-right: recent journal entries and correction alerts
- middle-left: open quests and collectible rewards
- middle-right: unrealized loot and appraisal totals
- bottom-left: write-off candidates and stale items
- bottom-right: quick navigation, hotkeys, or selected account summary

### Interaction Model

Suggested input model:

- single-key navigation for major sections
- arrow keys or `hjkl` for movement within lists
- enter to inspect or drill down
- explicit actions for `post`, `reverse`, `collect`, `write off`, and `recognize`
- mouse support as a convenience, not a requirement

## Storage Design

Planned database: SQLite

Planned top-level tables:

- `accounts`
- `journal_entries`
- `journal_lines`
- `quests`
- `quest_reward_terms`
- `quest_events`
- `loot_items`
- `loot_appraisals`
- `settings`

The initial storage engine should be SQLite.

Reasons:

- local-first workflow
- zero external infrastructure
- sufficient for single-user campaign bookkeeping
- easy backup and export behavior

Initialization data should be stored under `src/config` as setup assets:

- schema SQL for creating the first tables
- default seed data such as the initial chart of accounts

Those setup assets are consumed only during initialization of an empty database. They must not overwrite an existing LootSheet database on later runs.

### Accounts

Fields:

- `id`
- `code`
- `name`
- `type`
- `active`
- `created_at`
- `updated_at`

### Journal Entries

Fields:

- `id`
- `entry_number`
- `status`
- `entry_date`
- `description`
- `reverses_entry_id`
- `created_at`
- `posted_at`

### Journal Lines

Fields:

- `id`
- `journal_entry_id`
- `account_id`
- `memo`
- `debit_amount`
- `credit_amount`

### Quests

Fields:

- `id`
- `title`
- `patron`
- `status`
- `notes`
- `accepted_on`
- `completed_on`
- `closed_on`

### Quest Reward Terms

Fields:

- `id`
- `quest_id`
- `term_type`
- `amount`
- `condition_text`
- `earned`
- `paid`

### Loot Items

Fields:

- `id`
- `name`
- `source`
- `status`
- `quantity`
- `notes`

### Loot Appraisals

Fields:

- `id`
- `loot_item_id`
- `appraised_value`
- `appraiser`
- `notes`
- `appraised_at`
- `recognized_entry_id`

## Reporting

Reports needed in v1:

- trial balance
- account ledger
- open quests
- collectible rewards
- unrealized loot summary
- correction history

## Non-Goals

Not planned for v1:

- authentication
- cloud sync
- browser UI
- tax logic
- invoice workflows

## Open Questions

- Should each party member eventually have a personal subledger?
- Should consumables support optional quantity tracking without becoming full inventory accounting?
- Should appraisal recognition be encouraged, optional, or disabled by default?
- How much automation should be provided for common journal templates?
