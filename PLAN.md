# Plan

This file describes the intended implementation sequence for LootSheet.

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

Exit criteria:
- a user can initialize a local database
- a user can post balanced journal entries from the CLI
- a user can inspect accounts and registers from the CLI

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
- outstanding quest report
- unrealized loot report
- correction history

Exit criteria:
- a user can understand current cash, earned-but-unpaid rewards, and unrealized loot at a glance

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

## Non-Goals for v1

- online sync
- multiplayer access
- browser frontend
- tax, VAT, payroll, or invoice support
- perfect inventory simulation for every trivial item
