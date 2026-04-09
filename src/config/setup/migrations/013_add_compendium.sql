-- Compendium tables for cross-campaign D&D Beyond reference data.
-- No campaign_id — these are shared across all campaigns.

CREATE TABLE compendium_sources (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0, 1))
);

CREATE TABLE compendium_monsters (
    id          INTEGER PRIMARY KEY,
    ddb_id      INTEGER UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    cr          TEXT NOT NULL DEFAULT '',
    type        TEXT NOT NULL DEFAULT '',
    size        TEXT NOT NULL DEFAULT '',
    hp          TEXT NOT NULL DEFAULT '',
    ac          TEXT NOT NULL DEFAULT '',
    source_name TEXT NOT NULL DEFAULT '',
    detail_json TEXT NOT NULL DEFAULT '{}',
    synced_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE compendium_spells (
    id           INTEGER PRIMARY KEY,
    ddb_id       INTEGER UNIQUE NOT NULL,
    name         TEXT NOT NULL,
    level        INTEGER NOT NULL DEFAULT 0,
    school       TEXT NOT NULL DEFAULT '',
    casting_time TEXT NOT NULL DEFAULT '',
    range        TEXT NOT NULL DEFAULT '',
    components   TEXT NOT NULL DEFAULT '',
    duration     TEXT NOT NULL DEFAULT '',
    classes      TEXT NOT NULL DEFAULT '',
    source_name  TEXT NOT NULL DEFAULT '',
    detail_json  TEXT NOT NULL DEFAULT '{}',
    synced_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE compendium_items (
    id          INTEGER PRIMARY KEY,
    ddb_id      INTEGER UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL DEFAULT '',
    rarity      TEXT NOT NULL DEFAULT '',
    attunement  INTEGER NOT NULL DEFAULT 0 CHECK (attunement IN (0, 1)),
    source_name TEXT NOT NULL DEFAULT '',
    detail_json TEXT NOT NULL DEFAULT '{}',
    synced_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE compendium_rules (
    id          INTEGER PRIMARY KEY,
    ddb_id      INTEGER UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    category    TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    synced_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE compendium_conditions (
    id          INTEGER PRIMARY KEY,
    ddb_id      INTEGER UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    synced_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

-- FTS5 indexes for full-text search.

CREATE VIRTUAL TABLE compendium_monsters_fts USING fts5(
    name, type, source_name,
    content='compendium_monsters',
    content_rowid='id'
);

CREATE TRIGGER compendium_monsters_fts_insert AFTER INSERT ON compendium_monsters BEGIN
    INSERT INTO compendium_monsters_fts(rowid, name, type, source_name)
        VALUES (new.id, new.name, new.type, new.source_name);
END;
CREATE TRIGGER compendium_monsters_fts_delete AFTER DELETE ON compendium_monsters BEGIN
    INSERT INTO compendium_monsters_fts(compendium_monsters_fts, rowid, name, type, source_name)
        VALUES ('delete', old.id, old.name, old.type, old.source_name);
END;
CREATE TRIGGER compendium_monsters_fts_update AFTER UPDATE ON compendium_monsters BEGIN
    INSERT INTO compendium_monsters_fts(compendium_monsters_fts, rowid, name, type, source_name)
        VALUES ('delete', old.id, old.name, old.type, old.source_name);
    INSERT INTO compendium_monsters_fts(rowid, name, type, source_name)
        VALUES (new.id, new.name, new.type, new.source_name);
END;

CREATE VIRTUAL TABLE compendium_spells_fts USING fts5(
    name, school, classes, source_name,
    content='compendium_spells',
    content_rowid='id'
);

CREATE TRIGGER compendium_spells_fts_insert AFTER INSERT ON compendium_spells BEGIN
    INSERT INTO compendium_spells_fts(rowid, name, school, classes, source_name)
        VALUES (new.id, new.name, new.school, new.classes, new.source_name);
END;
CREATE TRIGGER compendium_spells_fts_delete AFTER DELETE ON compendium_spells BEGIN
    INSERT INTO compendium_spells_fts(compendium_spells_fts, rowid, name, school, classes, source_name)
        VALUES ('delete', old.id, old.name, old.school, old.classes, old.source_name);
END;
CREATE TRIGGER compendium_spells_fts_update AFTER UPDATE ON compendium_spells BEGIN
    INSERT INTO compendium_spells_fts(compendium_spells_fts, rowid, name, school, classes, source_name)
        VALUES ('delete', old.id, old.name, old.school, old.classes, old.source_name);
    INSERT INTO compendium_spells_fts(rowid, name, school, classes, source_name)
        VALUES (new.id, new.name, new.school, new.classes, new.source_name);
END;

CREATE VIRTUAL TABLE compendium_items_fts USING fts5(
    name, type, rarity, source_name,
    content='compendium_items',
    content_rowid='id'
);

CREATE TRIGGER compendium_items_fts_insert AFTER INSERT ON compendium_items BEGIN
    INSERT INTO compendium_items_fts(rowid, name, type, rarity, source_name)
        VALUES (new.id, new.name, new.type, new.rarity, new.source_name);
END;
CREATE TRIGGER compendium_items_fts_delete AFTER DELETE ON compendium_items BEGIN
    INSERT INTO compendium_items_fts(compendium_items_fts, rowid, name, type, rarity, source_name)
        VALUES ('delete', old.id, old.name, old.type, old.rarity, old.source_name);
END;
CREATE TRIGGER compendium_items_fts_update AFTER UPDATE ON compendium_items BEGIN
    INSERT INTO compendium_items_fts(compendium_items_fts, rowid, name, type, rarity, source_name)
        VALUES ('delete', old.id, old.name, old.type, old.rarity, old.source_name);
    INSERT INTO compendium_items_fts(rowid, name, type, rarity, source_name)
        VALUES (new.id, new.name, new.type, new.rarity, new.source_name);
END;

CREATE VIRTUAL TABLE compendium_rules_fts USING fts5(
    name, category,
    content='compendium_rules',
    content_rowid='id'
);

CREATE TRIGGER compendium_rules_fts_insert AFTER INSERT ON compendium_rules BEGIN
    INSERT INTO compendium_rules_fts(rowid, name, category)
        VALUES (new.id, new.name, new.category);
END;
CREATE TRIGGER compendium_rules_fts_delete AFTER DELETE ON compendium_rules BEGIN
    INSERT INTO compendium_rules_fts(compendium_rules_fts, rowid, name, category)
        VALUES ('delete', old.id, old.name, old.category);
END;
CREATE TRIGGER compendium_rules_fts_update AFTER UPDATE ON compendium_rules BEGIN
    INSERT INTO compendium_rules_fts(compendium_rules_fts, rowid, name, category)
        VALUES ('delete', old.id, old.name, old.category);
    INSERT INTO compendium_rules_fts(rowid, name, category)
        VALUES (new.id, new.name, new.category);
END;

CREATE VIRTUAL TABLE compendium_conditions_fts USING fts5(
    name,
    content='compendium_conditions',
    content_rowid='id'
);

CREATE TRIGGER compendium_conditions_fts_insert AFTER INSERT ON compendium_conditions BEGIN
    INSERT INTO compendium_conditions_fts(rowid, name)
        VALUES (new.id, new.name);
END;
CREATE TRIGGER compendium_conditions_fts_delete AFTER DELETE ON compendium_conditions BEGIN
    INSERT INTO compendium_conditions_fts(compendium_conditions_fts, rowid, name)
        VALUES ('delete', old.id, old.name);
END;
CREATE TRIGGER compendium_conditions_fts_update AFTER UPDATE ON compendium_conditions BEGIN
    INSERT INTO compendium_conditions_fts(compendium_conditions_fts, rowid, name)
        VALUES ('delete', old.id, old.name);
    INSERT INTO compendium_conditions_fts(rowid, name)
        VALUES (new.id, new.name);
END;
