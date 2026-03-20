package app

const rootHelpText = `LootSheet CLI

Usage:
  lootsheet COMMAND [SUBCOMMAND] [FLAGS]
  lootsheet help [COMMAND] [SUBCOMMAND]
  lootsheet COMMAND help
  lootsheet COMMAND SUBCOMMAND -h
  lootsheet COMMAND SUBCOMMAND --help

Commands:
  db         inspect database state and run schema migrations
  init       initialize a fresh LootSheet database
  tui        open the full-screen TUI shell (local database)
  serve      host the database over the network for remote TUI clients
  connect    join a remote LootSheet server and open the TUI

Examples:
  lootsheet init
  lootsheet tui
  lootsheet serve --addr :7547
  lootsheet connect localhost:7547
  lootsheet db status
  lootsheet db migrate

Help examples:
  lootsheet help
  lootsheet db help
  lootsheet db status --help
`

const dbHelpText = `LootSheet CLI

Usage:
  lootsheet db status
  lootsheet db migrate

Subcommands:
  status   inspect whether the configured SQLite database is uninitialized, current, upgradeable, foreign, or damaged
  migrate  apply pending embedded schema migrations to an existing LootSheet database

Examples:
  lootsheet db status
  lootsheet db migrate
`

const dbStatusHelpText = `LootSheet CLI

Usage:
  lootsheet db status

Displays:
  Database path
  Existence and lifecycle state
  Detail for foreign or damaged databases
  Current and target schema versions
  Applied and pending migrations
`

const dbMigrateHelpText = `LootSheet CLI

Usage:
  lootsheet db migrate

Applies pending schema migrations to the configured LootSheet database.
Foreign and damaged databases are reported and left untouched.

Examples:
  lootsheet db migrate
`

const initHelpText = `LootSheet CLI

Usage:
  lootsheet init

Bootstraps a fresh SQLite database from the embedded schema and seed accounts.
If the configured database already contains LootSheet metadata, init reports that it is already initialized and does not reseed it.
`

const serveHelpText = `LootSheet CLI

Usage:
  lootsheet serve [--addr HOST:PORT]

Hosts the configured LootSheet database over TLS for remote TUI clients.
On first start, generates a self-signed TLS certificate and a bearer token
in the data directory. The token is printed to stdout for sharing with
party members.

Flags:
  --addr   listen address (default :7547)

Examples:
  lootsheet serve
  lootsheet serve --addr 0.0.0.0:9000
`

const connectHelpText = `LootSheet CLI

Usage:
  lootsheet connect <addr>

Connects to a remote LootSheet server and opens the full TUI shell. On first
connection, prompts for the server bearer token. The token is saved in the
config directory for subsequent sessions.

Examples:
  lootsheet connect localhost:7547
  lootsheet connect 192.168.1.10:7547
`

const tuiHelpText = `LootSheet CLI

Usage:
  lootsheet tui

Opens the full-screen LootSheet TUI using tcell.
The current slice is interactive: a boxed dashboard plus Accounts, Journal, Quest, Loot, Assets, Codex, and Notes screens backed by app-facing adapters. List screens keep a selected row and detail pane visible, the shell redraws cleanly on resize, the dashboard exposes guided expense, income, and custom journal-entry launchers, the Accounts screen supports add/remove/toggle actions, the Journal screen exposes guided expense/income launchers plus reversal, the Quest screen supports add/edit plus collect/write-off actions, the Loot screen supports add/edit plus recognize/sell actions, the Codex screen supports add/edit/delete actions with type-specific forms, and the Notes screen supports add/edit/delete actions.

Keys:
  Left/Right, Tab/Shift+Tab  move between top-level sections
  1-8                        jump directly to dashboard/accounts/journal/quests/loot/assets/codex/notes
  Up/Down, j/k               move the selected row on list screens
  PgUp/PgDn, Home/End        jump through longer list screens
  e                          open guided expense entry creation on the Dashboard or Journal screen
  i                          open guided income entry creation on the Dashboard or Journal screen
  a                          open guided custom entry creation on the Dashboard; add accounts, quests, or loot on their screens
  u                          edit the selected quest or loot item on its screen
  d                          remove the selected account on the Accounts screen
  t                          toggle the selected account active/inactive on the Accounts screen
  r                          reverse the selected posted journal entry on the Journal screen
  c                          collect the full outstanding balance for the selected quest on the Quest screen
  w                          write off the full outstanding balance for the selected quest on the Quest screen
  n                          recognize the selected latest loot appraisal on the Loot screen
  s                          sell the selected recognized loot item on the Loot screen
  Loot create/edit           does not set value; appraisal happens later as a separate workflow
  ?                          open a glossary modal for the current screen's accounting terms
  Enter                      confirm the open modal
  Enter                      submit the guided entry composer when a guided entry form is open
  Esc                        cancel the open modal, or quit when no modal is open
  q                          cancel the open modal, or quit when no modal is open
  Ctrl+L                     reload data and force a full redraw
`
