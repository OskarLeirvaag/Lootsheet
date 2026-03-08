# AGENTS.md

Repo-specific rules for working in LootSheet.

## Read First

For meaningful product or architecture changes, read:

- [README.md](README.md)
- [DESIGN.md](DESIGN.md)
- [TODO.md](TODO.md)

Read [PLAN.md](PLAN.md) only when milestone sequencing matters.

## Invariants

Do not weaken these unless the user explicitly changes direction:

- posted journal entries are immutable
- journal entries must balance
- account IDs are immutable
- accounts with postings cannot be deleted
- used accounts may be marked inactive
- promised quest rewards stay off-ledger until earned
- unrealized loot appraisals stay off-ledger until explicitly recognized

## Repo Shape

Keep this structure:

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

Dependency flow stays one-way:

- `app -> domain packages -> ledger -> config`
- no cross-domain imports between `account`, `journal`, `quest`, `loot`, and `report`

## Engineering Preferences

- prefer simple explicit Go
- keep dependencies small
- prefer SQLite-compatible behavior
- preserve auditability over convenience
- keep docs and tests aligned with accounting behavior changes

## TUI Guidance

- boxed `btop`-style layout
- keyboard-first
- cell-based rendering
- no raw SQL in `src/render`
- mouse support is optional

## Quality

Preferred local gate:

- `make check`

## When Unsure

- prefer boring, easy-to-test designs
- do not introduce auth or network architecture unless explicitly requested
