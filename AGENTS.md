# AGENTS.md

Repo-specific rules for working in LootSheet.

## Read First

For meaningful product or architecture changes, read:

- [README.md](README.md)
- [TODO.md](TODO.md)

## Invariants

Do not weaken these unless the user explicitly changes direction:

- posted journal entries are immutable
- corrections happen through reversal, adjustment, or reclassification
- journal entries must balance
- account IDs are immutable
- accounts with postings cannot be deleted
- used accounts may be marked inactive
- promised quest rewards stay off-ledger until earned
- unrealized loot appraisals stay off-ledger until explicitly recognized

## Domain Rules

Account types: `asset`, `liability`, `equity`, `income`, `expense`. Custom accounts are a first-class feature.

Quest lifecycle: `offered` -> `accepted` -> `completed` -> `collectible` -> `partially_paid` / `paid` / `defaulted` / `voided`. Earned but unpaid rewards become receivables. Failed collection becomes a write-off, not silent deletion.

Loot lifecycle: create with no value -> appraise -> recognize on-ledger -> sell. Sale records gain or loss against the recognized basis.

Accounting assumptions: no VAT, invoice, or payroll. Small consumables default to expenses. Loot value is uncertain until sale.

## Repo Shape

Keep this structure:

- `main.go`
- `src/app` — CLI commands, TUI data loaders, command handlers
- `src/ledger` — SQLite storage, migrations, domain queries
- `src/ledger/export` — CSV, Excel, PDF export of trial balance and account ledgers
- `src/config` — configuration loading, embedded schema/seed assets
- `src/render` — TUI shell, views, compose forms, search modal, ledger overlay
- `src/net/proto` — protobuf definitions, version constants, model conversion
- `src/net/server` — TLS server for remote TUI sessions
- `src/net/client` — TLS client connecting to a remote server
- `src/net/wire` — length-prefixed message framing over TCP
- `src/currency` — GP/SP/CP parsing and formatting
- `src/texture` — shared text-processing helpers
- `src/testutil` — test helpers and fixtures

Dependency flow stays one-way:

- `app -> ledger -> config`
- `app -> render -> render/model`
- `app -> net/client -> net/proto -> render/model`
- `net/server -> net/proto -> render/model`
- `ledger/export -> ledger, currency`
- `currency` and `texture` are leaf packages with no internal dependencies

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

## Network Protocol

- client/server over TLS 1.3 with bearer token auth
- protobuf request/response over length-prefixed TCP frames
- version exchange during AUTH: `ProtocolVersion` (wire compat, hard reject on mismatch) + `AppVersion` (feature compat, checked against `MinServerVersion`/`MinClientVersion`)
- `ProtocolVersion` bumps only for wire-breaking changes; `AppVersion` bumps per release
- the `#` ledger overlay and export commands require server `AppVersion >= 0.6.0`

## When Unsure

- prefer boring, easy-to-test designs
