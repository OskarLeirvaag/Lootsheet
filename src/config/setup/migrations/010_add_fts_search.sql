-- FTS5 virtual tables for full-text search on notes and codex entries.
-- Uses content-sync triggers so the index stays current without Go code changes.

-- Notes FTS
CREATE VIRTUAL TABLE notes_fts USING fts5(
    title, body,
    content='notes',
    content_rowid='rowid'
);

-- Populate from existing data.
INSERT INTO notes_fts(rowid, title, body)
    SELECT rowid, title, body FROM notes;

-- Keep FTS in sync via triggers.
CREATE TRIGGER notes_fts_insert AFTER INSERT ON notes BEGIN
    INSERT INTO notes_fts(rowid, title, body) VALUES (new.rowid, new.title, new.body);
END;

CREATE TRIGGER notes_fts_delete AFTER DELETE ON notes BEGIN
    INSERT INTO notes_fts(notes_fts, rowid, title, body) VALUES ('delete', old.rowid, old.title, old.body);
END;

CREATE TRIGGER notes_fts_update AFTER UPDATE ON notes BEGIN
    INSERT INTO notes_fts(notes_fts, rowid, title, body) VALUES ('delete', old.rowid, old.title, old.body);
    INSERT INTO notes_fts(rowid, title, body) VALUES (new.rowid, new.title, new.body);
END;

-- Codex entries FTS
CREATE VIRTUAL TABLE codex_fts USING fts5(
    name, title, location, faction, notes, class, race, description,
    content='codex_entries',
    content_rowid='rowid'
);

-- Populate from existing data.
INSERT INTO codex_fts(rowid, name, title, location, faction, notes, class, race, description)
    SELECT rowid, name, title, location, faction, notes, class, race, description FROM codex_entries;

-- Keep FTS in sync via triggers.
CREATE TRIGGER codex_fts_insert AFTER INSERT ON codex_entries BEGIN
    INSERT INTO codex_fts(rowid, name, title, location, faction, notes, class, race, description)
        VALUES (new.rowid, new.name, new.title, new.location, new.faction, new.notes, new.class, new.race, new.description);
END;

CREATE TRIGGER codex_fts_delete AFTER DELETE ON codex_entries BEGIN
    INSERT INTO codex_fts(codex_fts, rowid, name, title, location, faction, notes, class, race, description)
        VALUES ('delete', old.rowid, old.name, old.title, old.location, old.faction, old.notes, old.class, old.race, old.description);
END;

CREATE TRIGGER codex_fts_update AFTER UPDATE ON codex_entries BEGIN
    INSERT INTO codex_fts(codex_fts, rowid, name, title, location, faction, notes, class, race, description)
        VALUES ('delete', old.rowid, old.name, old.title, old.location, old.faction, old.notes, old.class, old.race, old.description);
    INSERT INTO codex_fts(rowid, name, title, location, faction, notes, class, race, description)
        VALUES (new.rowid, new.name, new.title, new.location, new.faction, new.notes, new.class, new.race, new.description);
END;
