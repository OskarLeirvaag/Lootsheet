# Plan

This file describes the intended implementation sequence for LootSheet.

## Current Snapshot

- Phases 1 and 2 are largely complete: the SQLite-backed CLI workflows are in place.
- The CLI command tree now runs through Cobra, and the repository ships generated man pages from that command tree under `docs/man`.
- Several reporting deliverables from Phase 5 shipped early to support the CLI and smoke coverage.
- The next major product milestone is Phase 3, the `tcell`-based TUI shell.
- The main supporting work around that milestone is release packaging and installation polish.

## Phase 0: Product Definition

Goal:
- lock the accounting model and the scope for v1

Deliverables:
- README
- design document
- prioritized backlog
- contributor instructions

Exit criteria:
- the difference between `ledger` and `registers` is explicit
- accounting invariants are documented
- the first implementation target is local CLI/TUI

## Phase 1: Domain and Storage

Goal:
- define the core data structures and persist them locally

Deliverables:
- initial project layout with `main.go` and `src/...` packages
- Go module initialization
- SQLite schema and migrations
- repositories for accounts, journal entries, quests, and loot appraisals
- seed chart of accounts
- direct SQLite access from Go rather than shelling out to an external `sqlite3` executable

Exit criteria:
- application bootstrap flows through `main.go -> src/app`
- accounts can be created, renamed, and deactivated
- journal entries can be posted and validated
- quest and loot registers can be stored and queried

## Phase 2: CLI Foundation

Goal:
- provide a non-TUI interface for basic operations first

Deliverables:
- config file loading and local database bootstrap
- `lootsheet init`
- `lootsheet account`
- `lootsheet journal`
- `lootsheet quest`
- `lootsheet loot`
- `testapp.sh` smoke script that creates a temporary test workspace, initializes a test database, runs a representative CLI scenario, and verifies expected behavior and logs so an agent can execute it end to end

Exit criteria:
- a user can initialize a local database
- a user can post balanced journal entries from the CLI
- a user can inspect accounts and registers from the CLI
- an agent can run `testapp.sh` locally to validate the expected CLI flow and log output against a temporary database

## Phase 3: TUI Shell

Goal:
- build a navigable accounting-style terminal UI

Deliverables:
- `tcell` screen and renderer setup
- alternate-screen terminal lifecycle
- dashboard view
- account list view
- journal list view
- register summary widgets
- boxed panel layout inspired by `btop`
- theme and color groundwork
- shared navigation and help

Exit criteria:
- the app opens into a stable dashboard
- the panel layout resizes cleanly with terminal size changes
- a user can move between the main sections without using raw commands

## Phase 4: Core Workflows

Goal:
- support the workflows that make LootSheet useful in play

Deliverables:
- post journal entries
- reverse or adjust posted entries
- record quest promises and completion
- record loot appraisals and recognition
- write off failed receivables

Exit criteria:
- the main D&D accounting scenarios are covered end to end
- corrections happen without mutating posted data

## Phase 5: Reporting and Review

Goal:
- make the books understandable

Deliverables:
- trial balance
- account ledger drill-down
- quest receivables report
- promised-but-unearned quest report
- unrealized loot report
- write-off candidates report
- correction history later if it proves necessary for day-to-day review

Exit criteria:
- a user can understand current cash, promised rewards, earned-but-unpaid rewards, unrealized loot, and stale receivables at a glance

## Phase 6: Polish

Goal:
- make the tool pleasant enough to use regularly

Deliverables:
- import/export
- backups
- sample data
- keyboard shortcuts
- improved terminal styling

Exit criteria:
- the app is comfortable for weekly campaign use

## Phase 7: Packaging and Longevity

Goal:
- make LootSheet installable on real Linux and macOS systems and keep local data safe across upgrades over long-term use

Deliverables:
- release build targets for Linux and macOS first:
  - linux amd64
  - linux arm64
  - darwin amd64
  - darwin arm64
- Windows build target later if the SQLite, terminal, and packaging story remains clean enough
- installation documentation for:
  - direct binary install into a user-visible path
  - archive-based release installation
  - Homebrew formula or tap later if worth the maintenance
- storage/runtime layout documentation for:
  - config file
  - SQLite database
  - backup directory
  - optional export locations
- versioned SQLite migrations with explicit schema version checks
- startup behavior that clearly distinguishes:
  - first-time init
  - normal open of an existing database
  - upgrade requiring migration
  - damaged or foreign database detection
- backup and restore workflow suitable for local-first long-term use
- `testapp.sh` coverage extended to validate installed-binary behavior against a temporary workspace
- release checklist covering build, smoke test, migration safety, and packaging verification

Exit criteria:
- a user can install LootSheet on Linux or macOS without editing source files
- a user can upgrade without manually moving config or database files
- schema upgrades preserve existing data or fail with a clear recovery path
- the app stores config, database, backups, and logs in predictable per-user locations
- an agent can validate a release candidate with the smoke script against a temporary install/test workspace

## Runtime Model

The intended long-term runtime behavior should be:

- one local binary, no daemon, no background service
- one primary SQLite database per user-selected workspace or default per-user data directory
- config stored separately from accounting data
- backups created before risky migration or repair flows
- no required external `sqlite3` binary at runtime
- forward-only migrations with explicit schema version tracking
- human-readable logging to stderr by default, with future optional log-file support if needed
- graceful handling of app restarts, terminal closure, and interrupted commands without corrupting posted data

## Non-Goals for v1

- online sync
- multiplayer access
- browser frontend
- tax, VAT, payroll, or invoice support
- perfect inventory simulation for every trivial item
