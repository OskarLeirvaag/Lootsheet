ALTER TABLE loot_items ADD COLUMN item_type TEXT NOT NULL DEFAULT 'loot'
  CHECK (item_type IN ('loot', 'asset'));
