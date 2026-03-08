ALTER TABLE journal_entries ADD COLUMN reversed_at TEXT;

CREATE INDEX idx_journal_entries_reverses_entry_id ON journal_entries (reverses_entry_id);
