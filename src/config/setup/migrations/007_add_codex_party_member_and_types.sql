ALTER TABLE codex_entries ADD COLUMN party_member INTEGER NOT NULL DEFAULT 0;

UPDATE codex_entries SET party_member = 1 WHERE type_id = 'player';

INSERT OR IGNORE INTO codex_types (id, name, form_id) VALUES ('adversary', 'Adversary', 'npc');
INSERT OR IGNORE INTO codex_types (id, name, form_id) VALUES ('settlement', 'Settlement', 'settlement');
