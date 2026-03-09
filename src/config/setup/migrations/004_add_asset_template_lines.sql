CREATE TABLE asset_template_lines (
    id TEXT PRIMARY KEY,
    loot_item_id TEXT NOT NULL REFERENCES loot_items(id) ON DELETE CASCADE,
    side TEXT NOT NULL CHECK (side IN ('debit', 'credit')),
    account_code TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_asset_template_lines_item ON asset_template_lines (loot_item_id, sort_order);
