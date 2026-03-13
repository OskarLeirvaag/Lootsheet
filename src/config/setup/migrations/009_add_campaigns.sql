-- Create campaigns table
CREATE TABLE campaigns (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default campaign for existing data
INSERT INTO campaigns (id, name) VALUES ('default', 'Default');

-- Rebuild accounts: change UNIQUE(code) → UNIQUE(campaign_id, code)
CREATE TABLE accounts_new (
    id          TEXT PRIMARY KEY,
    campaign_id TEXT NOT NULL DEFAULT 'default' REFERENCES campaigns(id),
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('asset','liability','equity','income','expense')),
    active      INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    created_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (campaign_id, code)
);
INSERT INTO accounts_new SELECT id, 'default', code, name, type, active, created_at, updated_at FROM accounts;
DROP TABLE accounts;
ALTER TABLE accounts_new RENAME TO accounts;
CREATE INDEX idx_accounts_active ON accounts (active, name);
CREATE INDEX idx_accounts_campaign ON accounts (campaign_id);

-- Rebuild journal_entries: change UNIQUE(entry_number) → UNIQUE(campaign_id, entry_number)
CREATE TABLE journal_entries_new (
    id                TEXT PRIMARY KEY,
    campaign_id       TEXT NOT NULL DEFAULT 'default' REFERENCES campaigns(id),
    entry_number      INTEGER NOT NULL,
    status            TEXT NOT NULL CHECK (status IN ('draft','posted','reversed')),
    entry_date        TEXT NOT NULL,
    description       TEXT NOT NULL,
    reverses_entry_id TEXT REFERENCES journal_entries_new(id) ON DELETE RESTRICT,
    created_at        TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    posted_at         TEXT,
    reversed_at       TEXT,
    UNIQUE (campaign_id, entry_number)
);
INSERT INTO journal_entries_new SELECT id, 'default', entry_number, status, entry_date, description, reverses_entry_id, created_at, posted_at, reversed_at FROM journal_entries;
DROP TABLE journal_entries;
ALTER TABLE journal_entries_new RENAME TO journal_entries;
CREATE INDEX idx_journal_entries_status_date ON journal_entries (status, entry_date);
CREATE INDEX idx_journal_entries_reverses ON journal_entries (reverses_entry_id);
CREATE INDEX idx_journal_entries_campaign ON journal_entries (campaign_id);

-- Simple ALTER for remaining tables
ALTER TABLE quests ADD COLUMN campaign_id TEXT NOT NULL DEFAULT 'default' REFERENCES campaigns(id);
CREATE INDEX idx_quests_campaign ON quests (campaign_id);

ALTER TABLE loot_items ADD COLUMN campaign_id TEXT NOT NULL DEFAULT 'default' REFERENCES campaigns(id);
CREATE INDEX idx_loot_items_campaign ON loot_items (campaign_id);

ALTER TABLE codex_entries ADD COLUMN campaign_id TEXT NOT NULL DEFAULT 'default' REFERENCES campaigns(id);
CREATE INDEX idx_codex_entries_campaign ON codex_entries (campaign_id);

ALTER TABLE notes ADD COLUMN campaign_id TEXT NOT NULL DEFAULT 'default' REFERENCES campaigns(id);
CREATE INDEX idx_notes_campaign ON notes (campaign_id);

ALTER TABLE entity_references ADD COLUMN campaign_id TEXT NOT NULL DEFAULT 'default' REFERENCES campaigns(id);
CREATE INDEX idx_entity_references_campaign ON entity_references (campaign_id);

-- Track active campaign in settings
INSERT INTO settings (key, value) VALUES ('active_campaign_id', 'default');
