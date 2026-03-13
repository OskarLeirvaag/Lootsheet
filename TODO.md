# TODO

## Active

- [ ] Define release targets for Linux amd64/arm64 and macOS amd64/arm64.
- [ ] Decide the first install format.
- [ ] Add a `backup` command.
- [ ] Add CSV export.
- [ ] Add sample dataset import.
- [ ] Expand structured logging only if a real need appears.

## TUI

- [ ] Optional mouse support.

## Later

- [ ] Configurable GP/SP/CP display helpers.
- [ ] Optional member balance and distribution tracking.
- [ ] Evaluate Windows support after Linux/macOS packaging is stable.

## Already Done

- [x] Core SQLite storage and migrations.
- [x] CLI for init, database status, and migrations.
- [x] TUI shell with seven sections: Dashboard, Journal, Quests, Loot, Assets, Codex, Notes.
- [x] Compose forms for journal entries, quests, loot, assets, codex entries, and notes.
- [x] Quest and loot lifecycle workflows (earn, collect, write-off, appraise, recognize, sell).
- [x] Asset register with journal entry templates.
- [x] Codex and notes with entity cross-references.
- [x] Full-text search across all sections.
- [x] Glossary of accounting and game terms.
- [x] Upgrade detection, migration execution, and backup-before-risky-migration behavior.
