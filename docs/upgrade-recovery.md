# Upgrade And Recovery

This document describes the current supported database upgrade and recovery flow for LootSheet.

## Before You Upgrade

Check the database state first:

```sh
lootsheet db status
```

Possible states:

- `uninitialized`: no LootSheet database exists yet at the configured path
- `current`: the database matches the embedded schema version for this build
- `upgradeable`: the database is a LootSheet database and this build has later embedded migrations available
- `foreign`: the file is SQLite, but it does not look like a LootSheet database this build can safely manage
- `damaged`: the file exists, but SQLite metadata is unreadable or invalid

`lootsheet db status` also prints:

- the configured database path
- the current schema version when available
- the target embedded schema version for this build
- applied and pending migration counts
- detail text for `foreign` and `damaged` states

## Supported Upgrade Flow

When `lootsheet db status` reports `upgradeable`, run:

```sh
lootsheet db migrate
```

Current behavior:

- LootSheet loads the embedded migrations for the current build
- it creates a timestamped backup before applying schema changes or repairing legacy metadata
- it applies pending migrations in a single transaction
- it updates `schema_migrations` and the `settings.schema_version` marker
- it prints the backup path and the before/after schema versions

Default backup location:

- `${LOOTSHEET_DATA_DIR}/backups`

Optional overrides:

- `LOOTSHEET_BACKUP_DIR`
- `LOOTSHEET_DATA_DIR`
- `LOOTSHEET_DATABASE_PATH`

## When To Use `init`

Use:

```sh
lootsheet init
```

only when the database state is `uninitialized`.

`init` bootstraps a fresh LootSheet database from the embedded schema and seed accounts.
It is not the upgrade path for an existing LootSheet database.

## Foreign Or Damaged Databases

LootSheet does not auto-migrate `foreign` or `damaged` databases.

If `db status` reports `foreign`:

1. confirm that `LOOTSHEET_DATABASE_PATH` points at the intended file
2. confirm the file is actually a LootSheet database
3. if it is an older or partial database copy, restore the correct database file from backup instead of forcing migration

If `db status` reports `damaged`:

1. stop using that file as the primary database
2. locate the most recent known-good backup in the configured backup directory
3. copy the backup into place as the working database file
4. run `lootsheet db status` again before attempting `lootsheet db migrate`

## Restore Workflow

LootSheet does not yet have a dedicated `backup` or `restore` command.

Today, restore is a filesystem operation:

1. pick the backup file you want from the backup directory
2. copy it over the configured database path
3. rerun `lootsheet db status`
4. if the restored database is `upgradeable`, run `lootsheet db migrate`

## Related Assets

<<<<<<< Updated upstream
- generated CLI man pages: [`docs/man/`](docs/man)
- sample campaign fixture: [`fixtures/sample_campaign.sql`](fixtures/sample_campaign.sql)
=======
- generated CLI man pages: [`docs/man/`](man/)
- sample campaign fixture: [`fixtures/sample_campaign.sql`](../fixtures/sample_campaign.sql)
>>>>>>> Stashed changes
