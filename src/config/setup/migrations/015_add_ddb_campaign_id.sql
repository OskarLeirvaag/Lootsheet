-- Track the user's active D&D Beyond campaign so Phase B compendium syncs can
-- pass campaignId to spell/item fetches. With campaignId, DDB includes books
-- that have been *shared* into that campaign (e.g., a player gaining access to
-- "Keys from the Golden Vault" through their DM's content sharing). Without
-- it, only books the user purchased are returned.
--
-- 0 means "not set" — Phase A picks a sensible default (most-recently-created
-- DDB campaign) the first time it has cobalt.

ALTER TABLE compendium_sync_state
    ADD COLUMN ddb_campaign_id INTEGER NOT NULL DEFAULT 0;
