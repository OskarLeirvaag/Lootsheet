# Sample Fixtures

`sample_campaign.sql` is the canonical checked-in sample campaign dataset for LootSheet.

It is intentionally richer than the smoke test and includes:

- opening capital and normal expense activity
- a custom expense account (`5125 Tavern Reparations`)
- promised-but-unearned quests in both `offered` and `accepted` states
- a partially collected quest receivable old enough to surface in write-off candidate reports
- held, recognized, and sold loot with appraisal recognition and a loss-on-sale example
- codex entries for NPCs (Mayor Elra, Guild Factor Nera, Archivist Pell) and players (Ragnar, Mira)
- campaign notes (session recap, quest debrief, recon intel)
- entity references linking codex entries and notes to quests and people

The fixture is currently used for the demo GIF, regression tests, and documentation/examples.
There is not yet a user-facing `lootsheet import sample-dataset` workflow; that remains a later backlog item.
