-- Compendium ownership and per-source sync state.
-- Adds tri-state ownership flag, content-type heuristics, and last-synced
-- bookkeeping so the two-phase sync can avoid wasted DDB API calls.

ALTER TABLE compendium_sources ADD COLUMN owned       INTEGER NOT NULL DEFAULT 0 CHECK (owned IN (0, 1, 2));
ALTER TABLE compendium_sources ADD COLUMN has_spells  INTEGER NOT NULL DEFAULT 1 CHECK (has_spells IN (0, 1));
ALTER TABLE compendium_sources ADD COLUMN has_items   INTEGER NOT NULL DEFAULT 1 CHECK (has_items IN (0, 1));
ALTER TABLE compendium_sources ADD COLUMN is_released INTEGER NOT NULL DEFAULT 1 CHECK (is_released IN (0, 1));
ALTER TABLE compendium_sources ADD COLUMN category_id INTEGER NOT NULL DEFAULT 0;

-- Singleton row holding global sync state. id is always 1.
CREATE TABLE compendium_sync_state (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    last_synced_at  TEXT,
    last_phase      TEXT NOT NULL DEFAULT '',
    in_progress     INTEGER NOT NULL DEFAULT 0 CHECK (in_progress IN (0, 1))
);
INSERT INTO compendium_sync_state (id, last_synced_at, last_phase, in_progress) VALUES (1, NULL, '', 0);

-- Per-source sync timestamps and last error, used for resume hints.
CREATE TABLE compendium_source_sync (
    source_id          INTEGER PRIMARY KEY REFERENCES compendium_sources(id) ON DELETE CASCADE,
    monsters_synced_at TEXT,
    spells_synced_at   TEXT,
    items_synced_at    TEXT,
    last_error         TEXT NOT NULL DEFAULT ''
);
