CREATE TABLE entity_references (
    id          TEXT PRIMARY KEY,
    source_type TEXT NOT NULL CHECK (source_type IN ('codex','note','quest','loot')),
    source_id   TEXT NOT NULL,
    source_name TEXT NOT NULL DEFAULT '',
    target_type TEXT NOT NULL CHECK (target_type IN ('quest','loot','asset','person','note')),
    target_name TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_entity_references_source ON entity_references (source_type, source_id);
CREATE INDEX idx_entity_references_target ON entity_references (target_type, target_name);

-- Migrate existing data from codex_references
INSERT INTO entity_references (id, source_type, source_id, source_name, target_type, target_name, created_at)
SELECT cr.id, 'codex', cr.entry_id, ce.name, cr.target_type, cr.target_name, cr.created_at
FROM codex_references cr JOIN codex_entries ce ON ce.id = cr.entry_id;

-- Migrate existing data from notes_references
INSERT INTO entity_references (id, source_type, source_id, source_name, target_type, target_name, created_at)
SELECT nr.id, 'note', nr.note_id, n.title, nr.target_type, nr.target_name, nr.created_at
FROM notes_references nr JOIN notes n ON n.id = nr.note_id;

DROP TABLE codex_references;
DROP TABLE notes_references;
