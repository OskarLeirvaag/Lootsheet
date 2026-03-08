PRAGMA foreign_keys = ON;

CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('asset', 'liability', 'equity', 'income', 'expense')),
    active INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE journal_entries (
    id TEXT PRIMARY KEY,
    entry_number INTEGER UNIQUE,
    status TEXT NOT NULL CHECK (status IN ('draft', 'posted', 'reversed')),
    entry_date TEXT NOT NULL,
    description TEXT NOT NULL,
    reverses_entry_id TEXT REFERENCES journal_entries(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    posted_at TEXT
);

CREATE TABLE journal_lines (
    id TEXT PRIMARY KEY,
    journal_entry_id TEXT NOT NULL REFERENCES journal_entries(id) ON DELETE RESTRICT,
    line_number INTEGER NOT NULL,
    account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    memo TEXT NOT NULL DEFAULT '',
    debit_amount INTEGER NOT NULL DEFAULT 0 CHECK (debit_amount >= 0),
    credit_amount INTEGER NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
    CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR
        (credit_amount > 0 AND debit_amount = 0)
    ),
    UNIQUE (journal_entry_id, line_number)
);

CREATE TABLE quests (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    patron TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    promised_base_reward INTEGER NOT NULL DEFAULT 0 CHECK (promised_base_reward >= 0),
    partial_advance INTEGER NOT NULL DEFAULT 0 CHECK (partial_advance >= 0),
    bonus_conditions TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN (
        'offered',
        'accepted',
        'completed',
        'collectible',
        'partially_paid',
        'paid',
        'defaulted',
        'voided'
    )),
    notes TEXT NOT NULL DEFAULT '',
    accepted_on TEXT,
    completed_on TEXT,
    closed_on TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE loot_items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN (
        'held',
        'recognized',
        'sold',
        'assigned',
        'consumed',
        'discarded'
    )),
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    holder TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE loot_appraisals (
    id TEXT PRIMARY KEY,
    loot_item_id TEXT NOT NULL REFERENCES loot_items(id) ON DELETE RESTRICT,
    appraised_value INTEGER NOT NULL CHECK (appraised_value >= 0),
    appraiser TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    appraised_at TEXT NOT NULL,
    recognized_entry_id TEXT REFERENCES journal_entries(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounts_active ON accounts (active, name);
CREATE INDEX idx_journal_entries_status_date ON journal_entries (status, entry_date);
CREATE INDEX idx_journal_lines_account_id ON journal_lines (account_id);
CREATE INDEX idx_quests_status ON quests (status);
CREATE INDEX idx_loot_items_status ON loot_items (status);
