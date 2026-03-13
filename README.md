# LootSheet

[![Go](https://img.shields.io/github/go-mod/go-version/OskarLeirvaag/Lootsheet)](https://go.dev/)
[![License: GPL v3](https://img.shields.io/github/license/OskarLeirvaag/Lootsheet)](LICENSE)
[![Top Language](https://img.shields.io/github/languages/top/OskarLeirvaag/Lootsheet)](https://github.com/OskarLeirvaag/Lootsheet)
[![Code Size](https://img.shields.io/github/languages/code-size/OskarLeirvaag/Lootsheet)](https://github.com/OskarLeirvaag/Lootsheet)

LootSheet is a local bookkeeping tool for D&D 5e parties.

It is meant for groups that want to track party money, quest rewards, loot, and expenses with a little more rigor than a notes app or spreadsheet.

![LootSheet TUI demo](docs/demo.gif)

## What It Is For

LootSheet is for things like:

- recording shared party expenses such as arrows, rations, repairs, inns, and spell supplies
- tracking promised quest rewards before they are actually earned
- tracking loot and party assets before they are sold or disposed of
- keeping a codex of NPCs, players, and contacts encountered in the campaign
- writing session notes with cross-references to quests, loot, and people
- seeing what the party has in cash, what is still owed, and what has already been recognized in the books
- keeping a clean history instead of rewriting old entries

## What It Tries To Feel Like

LootSheet is accounting software first and a fantasy tool second.

That means:

- posted journal history is meant to stay intact
- mistakes are corrected, not erased
- quest promises and loot appraisals can stay off the books until they become real accounting events
- custom account names are welcome, including silly ones

## Features

The TUI provides seven main sections:

| Section | Description |
|---------|-------------|
| **Dashboard** | Overview of party finances, recent activity, and quick entry |
| **Journal** | Immutable double-entry ledger with reversal support |
| **Quests** | Quest register tracking rewards from offered through paid or defaulted |
| **Loot** | Unrealized loot register with appraisal, recognition, and sale workflows |
| **Assets** | Party asset register with journal entry templates |
| **Codex** | In-game reference book for NPCs, players, and contacts |
| **Notes** | Campaign and session notes with cross-references |

Additional capabilities:

- full-text search across all sections
- glossary of accounting and game terms
- configurable chart of accounts
- GP/SP/CP currency formatting throughout

## Interface

LootSheet is used primarily through a full-screen terminal TUI (`lootsheet tui`).

A small CLI surface handles setup and database management:

- `lootsheet init` — bootstrap a fresh database
- `lootsheet db status` — inspect database lifecycle state
- `lootsheet db migrate` — apply pending schema migrations

## Intended Use

Use LootSheet if your group wants a shared record of:

- party cash
- income from quests and rewards
- expenses and supply purchases
- loot found, appraised, recognized, and sold
- party assets and their accounting templates
- NPCs, contacts, and session notes
- what is still collectible and what has already been written off

It is not meant to be:

- a web app
- a multiplayer sync service
- a full inventory simulator
- a tax or invoice system

## Tone

The product should feel credible to someone who understands bookkeeping, while still fitting a D&D campaign where accounts like `Wizard Magic Ink`, `Arrows & Ammunition`, or `Tavern Reparations` make perfect sense.

## License

LootSheet is licensed under the [GNU General Public License v3.0](LICENSE).
