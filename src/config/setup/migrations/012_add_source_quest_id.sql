-- Add source_quest_id to journal_entries for stable quest payment tracking.
-- Replaces the fragile memo-text matching pattern.
ALTER TABLE journal_entries ADD COLUMN source_quest_id TEXT REFERENCES quests(id);
