-- Rename tables
ALTER TABLE people RENAME TO codex_entries;
ALTER TABLE people_references RENAME TO codex_references;

-- Rename foreign key column in references
ALTER TABLE codex_references RENAME COLUMN person_id TO entry_id;

-- Type system
CREATE TABLE codex_types (
    id      TEXT PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    form_id TEXT NOT NULL DEFAULT 'npc'
);

INSERT INTO codex_types (id, name, form_id) VALUES ('player', 'Player', 'player');
INSERT INTO codex_types (id, name, form_id) VALUES ('npc', 'NPC', 'npc');

-- New columns on codex_entries
ALTER TABLE codex_entries ADD COLUMN type_id     TEXT NOT NULL DEFAULT 'npc' REFERENCES codex_types(id);
ALTER TABLE codex_entries ADD COLUMN class       TEXT NOT NULL DEFAULT '';
ALTER TABLE codex_entries ADD COLUMN race        TEXT NOT NULL DEFAULT '';
ALTER TABLE codex_entries ADD COLUMN background  TEXT NOT NULL DEFAULT '';
ALTER TABLE codex_entries ADD COLUMN description TEXT NOT NULL DEFAULT '';

-- Migrate existing party members to player type
UPDATE codex_entries SET type_id = 'player' WHERE party_member = 1;

-- Recreate indexes with new names
DROP INDEX IF EXISTS idx_people_references_person;
DROP INDEX IF EXISTS idx_people_references_target;
CREATE INDEX idx_codex_references_entry ON codex_references (entry_id);
CREATE INDEX idx_codex_references_target ON codex_references (target_type, target_name);
