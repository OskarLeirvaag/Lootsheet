-- Add player_name column to codex_entries for tracking which real person
-- plays a character. Only meaningful for player-type entries.
ALTER TABLE codex_entries ADD COLUMN player_name TEXT NOT NULL DEFAULT '';

-- Rebuild codex FTS to include the new column.
DROP TRIGGER IF EXISTS codex_fts_insert;
DROP TRIGGER IF EXISTS codex_fts_delete;
DROP TRIGGER IF EXISTS codex_fts_update;
DROP TABLE IF EXISTS codex_fts;

CREATE VIRTUAL TABLE codex_fts USING fts5(
    name, title, location, faction, notes, class, race, description, player_name,
    content='codex_entries',
    content_rowid='rowid'
);

INSERT INTO codex_fts(rowid, name, title, location, faction, notes, class, race, description, player_name)
    SELECT rowid, name, title, location, faction, notes, class, race, description, player_name FROM codex_entries;

CREATE TRIGGER codex_fts_insert AFTER INSERT ON codex_entries BEGIN
    INSERT INTO codex_fts(rowid, name, title, location, faction, notes, class, race, description, player_name)
        VALUES (new.rowid, new.name, new.title, new.location, new.faction, new.notes, new.class, new.race, new.description, new.player_name);
END;

CREATE TRIGGER codex_fts_delete AFTER DELETE ON codex_entries BEGIN
    INSERT INTO codex_fts(codex_fts, rowid, name, title, location, faction, notes, class, race, description, player_name)
        VALUES ('delete', old.rowid, old.name, old.title, old.location, old.faction, old.notes, old.class, old.race, old.description, old.player_name);
END;

CREATE TRIGGER codex_fts_update AFTER UPDATE ON codex_entries BEGIN
    INSERT INTO codex_fts(codex_fts, rowid, name, title, location, faction, notes, class, race, description, player_name)
        VALUES ('delete', old.rowid, old.name, old.title, old.location, old.faction, old.notes, old.class, old.race, old.description, old.player_name);
    INSERT INTO codex_fts(rowid, name, title, location, faction, notes, class, race, description, player_name)
        VALUES (new.rowid, new.name, new.title, new.location, new.faction, new.notes, new.class, new.race, new.description, new.player_name);
END;
