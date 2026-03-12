-- Codex type system
CREATE TABLE codex_types (
    id      TEXT PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    form_id TEXT NOT NULL DEFAULT 'npc'
);

INSERT INTO codex_types (id, name, form_id) VALUES ('player', 'Player', 'player');
INSERT INTO codex_types (id, name, form_id) VALUES ('npc', 'NPC', 'npc');

-- Codex entries
CREATE TABLE codex_entries (
    id          TEXT PRIMARY KEY,
    type_id     TEXT NOT NULL DEFAULT 'npc' REFERENCES codex_types(id),
    name        TEXT NOT NULL,
    title       TEXT NOT NULL DEFAULT '',
    location    TEXT NOT NULL DEFAULT '',
    faction     TEXT NOT NULL DEFAULT '',
    disposition TEXT NOT NULL DEFAULT '',
    class       TEXT NOT NULL DEFAULT '',
    race        TEXT NOT NULL DEFAULT '',
    background  TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Codex cross-references
CREATE TABLE codex_references (
    id          TEXT PRIMARY KEY,
    entry_id    TEXT NOT NULL REFERENCES codex_entries(id) ON DELETE CASCADE,
    target_type TEXT NOT NULL CHECK (target_type IN ('quest', 'loot', 'asset', 'person')),
    target_name TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_codex_references_entry ON codex_references (entry_id);
CREATE INDEX idx_codex_references_target ON codex_references (target_type, target_name);

-- Notes
CREATE TABLE notes (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    body       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Notes cross-references
CREATE TABLE notes_references (
    id          TEXT PRIMARY KEY,
    note_id     TEXT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    target_type TEXT NOT NULL CHECK (target_type IN ('quest', 'loot', 'asset', 'person', 'note')),
    target_name TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notes_references_note ON notes_references (note_id);
CREATE INDEX idx_notes_references_target ON notes_references (target_type, target_name);
