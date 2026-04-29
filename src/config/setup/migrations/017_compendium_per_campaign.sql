-- Move compendium source selection from a global flag on compendium_sources
-- to a per-campaign junction table. Two campaigns can now enable different
-- source books (e.g., one uses MM 2014, another MM 2024).
--
-- Changes:
--   1. compendium_campaign_sources  — per-campaign (enabled, has_spells, has_items)
--   2. compendium_sources rebuilt   — drops the per-campaign columns
--   3. compendium_campaign_sync     — replaces the singleton compendium_sync_state
--                                    so each campaign tracks its own sync state
--
-- PRAGMA foreign_keys = OFF is handled by the migration runner.

-- 1. Per-campaign source selection.
CREATE TABLE compendium_campaign_sources (
    campaign_id TEXT    NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    source_id   INTEGER NOT NULL REFERENCES compendium_sources(id) ON DELETE CASCADE,
    enabled     INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0, 1)),
    has_spells  INTEGER NOT NULL DEFAULT 1 CHECK (has_spells IN (0, 1)),
    has_items   INTEGER NOT NULL DEFAULT 1 CHECK (has_items IN (0, 1)),
    PRIMARY KEY (campaign_id, source_id)
);

-- Migrate enabled sources for every existing campaign (CROSS JOIN so each
-- campaign starts with the same selection that was previously global).
INSERT INTO compendium_campaign_sources (campaign_id, source_id, enabled, has_spells, has_items)
SELECT c.id, cs.id, cs.enabled, cs.has_spells, cs.has_items
FROM campaigns c
CROSS JOIN compendium_sources cs
WHERE cs.enabled = 1 OR cs.has_spells = 0 OR cs.has_items = 0;

-- 2. Rebuild compendium_sources to keep only global/account-level columns.
--    Drops: enabled, has_spells, has_items (moved to junction above).
--    Keeps: id, name, owned, shared, is_released, category_id.
CREATE TABLE compendium_sources_new (
    id          INTEGER PRIMARY KEY,
    name        TEXT    NOT NULL,
    owned       INTEGER NOT NULL DEFAULT 0 CHECK (owned IN (0, 1, 2)),
    shared      INTEGER NOT NULL DEFAULT 0 CHECK (shared IN (0, 1)),
    is_released INTEGER NOT NULL DEFAULT 1 CHECK (is_released IN (0, 1)),
    category_id INTEGER NOT NULL DEFAULT 0
);
INSERT INTO compendium_sources_new (id, name, owned, shared, is_released, category_id)
    SELECT id, name, owned, shared, is_released, category_id FROM compendium_sources;
DROP TABLE compendium_sources;
ALTER TABLE compendium_sources_new RENAME TO compendium_sources;

-- 3. Per-campaign sync state (replaces the singleton compendium_sync_state).
CREATE TABLE compendium_campaign_sync (
    campaign_id     TEXT    NOT NULL PRIMARY KEY REFERENCES campaigns(id) ON DELETE CASCADE,
    last_synced_at  TEXT,
    ddb_campaign_id INTEGER NOT NULL DEFAULT 0,
    last_phase      TEXT    NOT NULL DEFAULT '',
    in_progress     INTEGER NOT NULL DEFAULT 0 CHECK (in_progress IN (0, 1))
);

-- Migrate existing sync state row (id=1) for all campaigns.
INSERT INTO compendium_campaign_sync (campaign_id, last_synced_at, ddb_campaign_id, last_phase, in_progress)
SELECT c.id, css.last_synced_at, css.ddb_campaign_id, css.last_phase, css.in_progress
FROM campaigns c
CROSS JOIN compendium_sync_state css WHERE css.id = 1;

DROP TABLE compendium_sync_state;

-- compendium_source_sync has no campaign_id column; leave it as-is for now
-- (the per-source error log is advisory; will be extended in a later migration).
