# Design

This file is the short source of truth for product invariants and domain rules.

## Product Shape

- local Go application
- CLI and TUI only
- SQLite as the system of record
- no auth, daemon, or required network dependency in v1

## Design Goals

- feel credible as accounting software
- keep the ledger auditable
- separate formal ledger records from operational registers
- support D&D-specific edge cases without weakening the bookkeeping model

## Architecture

Repository layout:

- `main.go`
- `src/app`
- `src/account`
- `src/journal`
- `src/quest`
- `src/loot`
- `src/report`
- `src/ledger`
- `src/config`
- `src/tools`
- `src/render`

Dependency flow:

- `app -> domain packages -> ledger -> config`
- `render` depends on app-facing data, not raw SQL
- no direct imports between `account`, `journal`, `quest`, `loot`, and `report`

## Hard Invariants

- posted journal entries are immutable
- corrections happen through reversal, adjustment, or reclassification
- journal entries must balance
- account IDs are immutable
- account names are editable
- accounts with postings cannot be deleted
- used accounts may be marked inactive

## Ledger vs Registers

The ledger is the formal accounting record.

Registers track operational state that may later create ledger entries.

Initial registers:

- quest register
- unrealized loot register

## Accounting Assumptions

- no VAT, invoice, or payroll model in v1
- small consumables default to expenses
- loot value is uncertain until sale
- a recognized appraisal can create visible gain or loss on later sale

## Account Types

Supported account types:

- `asset`
- `liability`
- `equity`
- `income`
- `expense`

Custom accounts are a first-class feature.

## Quest Rules

Each quest tracks:

- title
- patron
- description
- promised base reward
- optional partial advance
- optional bonus conditions
- status
- notes

Lifecycle statuses:

- `offered`
- `accepted`
- `completed`
- `collectible`
- `partially_paid`
- `paid`
- `defaulted`
- `voided`

Recognition rules:

- promised rewards stay off-ledger until earned
- earned but unpaid rewards become receivables
- failed collection becomes a write-off, not silent deletion

## Loot Rules

Each loot item tracks:

- name
- source
- quantity
- holder
- notes
- appraisals
- status

Lifecycle idea:

- create loot with no value
- appraise later in the loot register
- recognize only when the party chooses to move it on-ledger
- sale records gain or loss against the recognized basis

## TUI Direction

- keyboard-first
- boxed panel layout
- cell-based rendering
- section-specific color identity is fine
- mouse support is optional

## Storage and Operations

- per-user config/data locations
- forward-only migrations
- explicit database states: uninitialized, current, upgradeable, foreign, damaged
- backup before risky migration or repair work

For release planning and current backlog, see [PLAN.md](/home/raknoel/Projects/Private/Lootsheet/PLAN.md) and [TODO.md](/home/raknoel/Projects/Private/Lootsheet/TODO.md).
