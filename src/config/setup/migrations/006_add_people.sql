CREATE TABLE people (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    faction TEXT NOT NULL DEFAULT '',
    disposition TEXT NOT NULL DEFAULT '',
    party_member INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE people_references (
    id TEXT PRIMARY KEY,
    person_id TEXT NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    target_type TEXT NOT NULL CHECK (target_type IN ('quest', 'loot', 'asset', 'person')),
    target_name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_people_references_person ON people_references (person_id);
CREATE INDEX idx_people_references_target ON people_references (target_type, target_name);
