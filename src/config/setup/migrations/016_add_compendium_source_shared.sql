-- Adds a `shared` flag to compendium_sources so the picker can distinguish
-- "owned", "shared", and "locked" — rather than the previous binary
-- owned/not-owned model that hid campaign-shared content as if it were
-- inaccessible.
--
-- Phase A still sets `owned` from /mobile/api/v6/available-user-content.
-- Phase B observes which not-owned sources actually returned content (i.e.,
-- accessible via the user's selected DDB campaign) and flips `shared = 1` on
-- those. Display logic: owned > shared > locked > unknown.

ALTER TABLE compendium_sources
    ADD COLUMN shared INTEGER NOT NULL DEFAULT 0 CHECK (shared IN (0, 1));
