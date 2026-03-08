# Sample Fixtures

`sample_campaign.sql` is the canonical checked-in sample campaign dataset for LootSheet.

It is intentionally richer than the smoke test and includes:

- opening capital and normal expense activity
- a custom expense account (`5125 Tavern Reparations`)
- promised-but-unearned quests in both `offered` and `accepted` states
- a partially collected quest receivable old enough to surface in write-off candidate reports
- held, recognized, and sold loot with appraisal recognition and a loss-on-sale example

The fixture is currently used for regression tests and documentation/examples.
There is not yet a user-facing `lootsheet import sample-dataset` workflow; that remains a later backlog item.
