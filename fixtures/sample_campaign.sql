-- Sample campaign fixture for the Frostfall Campaign.
-- Designed to run on a freshly initialized database (lootsheet init).
-- Uses the 'default' campaign created by migration 009 and renames it.
-- Account IDs come from the init seed (party_cash, quest_receivable, etc.).

BEGIN TRANSACTION;

-- Rename the default campaign.
UPDATE campaigns SET name = 'Frostfall Campaign', created_at = '2026-02-01 07:00:00', updated_at = '2026-02-01 07:00:00' WHERE id = 'default';

-- Add a tavern reparations account (not in the default seed).
INSERT INTO accounts (id, campaign_id, code, name, type, active, created_at, updated_at)
VALUES ('tavern_reparations', 'default', '5125', 'Tavern Reparations', 'expense', 1, '2026-02-01 09:00:00', '2026-02-01 09:00:00');

-- ── Journal entries ──────────────────────────────────────────────────────

INSERT INTO journal_entries (id, campaign_id, entry_number, status, entry_date, description, reverses_entry_id, created_at, posted_at, reversed_at)
VALUES
	('je01', 'default', 1, 'posted', '2026-02-01', 'Party capital for the Frostfall campaign', NULL, '2026-02-01 09:00:00', '2026-02-01 09:00:00', NULL),
	('je02', 'default', 2, 'posted', '2026-02-02', 'Pay for broken inn chandelier', NULL, '2026-02-02 10:15:00', '2026-02-02 10:15:00', NULL),
	('je03', 'default', 3, 'posted', '2026-02-03', 'Restock arrows and chalk', NULL, '2026-02-03 08:30:00', '2026-02-03 08:30:00', NULL),
	('je04', 'default', 4, 'posted', '2026-02-05', 'Two nights at the Silver Oak Inn', NULL, '2026-02-05 07:00:00', '2026-02-05 07:00:00', NULL),
	('je05', 'default', 5, 'posted', '2026-02-07', 'Wizard ink for scroll copying', NULL, '2026-02-07 11:30:00', '2026-02-07 11:30:00', NULL),
	('je06', 'default', 6, 'posted', '2026-02-10', 'Moonlit Escort earned', NULL, '2026-02-10 18:20:00', '2026-02-10 18:20:00', NULL),
	('je07', 'default', 7, 'posted', '2026-02-15', 'Guild factor sent the first pouch', NULL, '2026-02-15 12:45:00', '2026-02-15 12:45:00', NULL),
	('je08', 'default', 8, 'posted', '2026-02-16', 'Healing potions and rations for the marsh', NULL, '2026-02-16 10:00:00', '2026-02-16 10:00:00', NULL),
	('je09', 'default', 9, 'posted', '2026-02-18', 'Recognize appraisal: Emerald Idol', NULL, '2026-02-18 14:00:00', '2026-02-18 14:00:00', NULL),
	('je10', 'default', 10, 'posted', '2026-02-19', 'Recognize appraisal: Cracked Ruby Crown', NULL, '2026-02-19 09:10:00', '2026-02-19 09:10:00', NULL),
	('je11', 'default', 11, 'posted', '2026-02-20', 'Sold cracked ruby crown below appraisal', NULL, '2026-02-20 16:30:00', '2026-02-20 16:30:00', NULL),
	('je12', 'default', 12, 'posted', '2026-02-22', 'Room and board at Frostfall Keep', NULL, '2026-02-22 08:00:00', '2026-02-22 08:00:00', NULL),
	('je13', 'default', 13, 'posted', '2026-02-24', 'Arrow restock at the keep armoury', NULL, '2026-02-24 09:30:00', '2026-02-24 09:30:00', NULL),
	('je14', 'default', 14, 'posted', '2026-03-01', 'Archivist escort advance from Pell', NULL, '2026-03-01 08:30:00', '2026-03-01 08:30:00', NULL);

INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount)
VALUES
	-- #1 Opening capital 150 GP
	('jl01a', 'je01', 1, 'party_cash', 'Initial purse and guild float', 15000, 0),
	('jl01b', 'je01', 2, 'party_equity', 'Initial purse and guild float', 0, 15000),
	-- #2 Tavern damage 3 GP 5 SP
	('jl02a', 'je02', 1, 'tavern_reparations', 'Silver Oak Inn settlement', 350, 0),
	('jl02b', 'je02', 2, 'party_cash', 'Silver Oak Inn settlement', 0, 350),
	-- #3 Arrow restock 1 GP 7 SP 5 CP
	('jl03a', 'je03', 1, 'arrows_and_ammunition', 'Arrows and chalk', 175, 0),
	('jl03b', 'je03', 2, 'party_cash', 'Arrows and chalk', 0, 175),
	-- #4 Inn 2 GP
	('jl04a', 'je04', 1, 'inn_and_travel', 'Two nights Silver Oak', 200, 0),
	('jl04b', 'je04', 2, 'party_cash', 'Two nights Silver Oak', 0, 200),
	-- #5 Wizard ink 4 GP
	('jl05a', 'je05', 1, 'wizard_magic_ink', 'Marsh fire ink', 400, 0),
	('jl05b', 'je05', 2, 'party_cash', 'Marsh fire ink', 0, 400),
	-- #6 Moonlit Escort earned 12 GP
	('jl06a', 'je06', 1, 'quest_receivable', 'Quest completion: Moonlit Escort', 1200, 0),
	('jl06b', 'je06', 2, 'quest_income', 'Quest completion: Moonlit Escort', 0, 1200),
	-- #7 Moonlit partial payment 7 GP
	('jl07a', 'je07', 1, 'party_cash', 'Quest payment: Moonlit Escort', 700, 0),
	('jl07b', 'je07', 2, 'quest_receivable', 'Quest payment: Moonlit Escort', 0, 700),
	-- #8 Supplies 2 GP 5 SP
	('jl08a', 'je08', 1, 'adventuring_supplies', 'Healing potions and rations', 250, 0),
	('jl08b', 'je08', 2, 'party_cash', 'Healing potions and rations', 0, 250),
	-- #9 Emerald Idol recognition 8 GP
	('jl09a', 'je09', 1, 'loot_inventory', 'Recognized: Emerald Idol', 800, 0),
	('jl09b', 'je09', 2, 'unrealized_loot_gain', 'Recognized: Emerald Idol', 0, 800),
	-- #10 Ruby Crown recognition 6 GP
	('jl10a', 'je10', 1, 'loot_inventory', 'Recognized: Cracked Ruby Crown', 600, 0),
	('jl10b', 'je10', 2, 'unrealized_loot_gain', 'Recognized: Cracked Ruby Crown', 0, 600),
	-- #11 Ruby Crown sale 4 GP (2 GP loss)
	('jl11a', 'je11', 1, 'party_cash', 'Sale proceeds: Cracked Ruby Crown', 400, 0),
	('jl11b', 'je11', 2, 'loss_on_sale_of_loot', 'Loss on sale: Cracked Ruby Crown', 200, 0),
	('jl11c', 'je11', 3, 'loot_inventory', 'Remove inventory: Cracked Ruby Crown', 0, 600),
	-- #12 Inn at Frostfall Keep 3 GP
	('jl12a', 'je12', 1, 'inn_and_travel', 'Room and board at the keep', 300, 0),
	('jl12b', 'je12', 2, 'party_cash', 'Room and board at the keep', 0, 300),
	-- #13 Arrow restock 1 GP 2 SP
	('jl13a', 'je13', 1, 'arrows_and_ammunition', 'Keep armoury', 120, 0),
	('jl13b', 'je13', 2, 'party_cash', 'Keep armoury', 0, 120),
	-- #14 Archivist escort advance 3 GP
	('jl14a', 'je14', 1, 'party_cash', 'Archivist escort advance', 300, 0),
	('jl14b', 'je14', 2, 'quest_income', 'Archivist escort advance', 0, 300);

-- ── Quests ───────────────────────────────────────────────────────────────

INSERT INTO quests (
	id, campaign_id, title, patron, description, promised_base_reward, partial_advance, bonus_conditions,
	status, notes, accepted_on, completed_on, closed_on, created_at, updated_at
)
VALUES
	('quest_clear_watchtower', 'default', 'Clear the Old Watchtower', 'Mayor Elra', 'Drive out the goblin squatters and recover the town bell.', 2000, 0, 'Bonus 5 GP if the bell is recovered intact', 'offered', '', NULL, NULL, NULL, '2026-02-25 09:00:00', '2026-02-25 09:00:00'),
	('quest_archivist_escort', 'default', 'Escort the Archivist', 'Archivist Pell', 'Protect the archivist on the road to Frostfall Keep.', 1500, 300, 'Extra 2 GP if no pages are singed', 'accepted', '', '2026-03-01', NULL, NULL, '2026-03-01 08:00:00', '2026-03-01 08:00:00'),
	('quest_moonlit_escort', 'default', 'Moonlit Escort', 'Guild Factor Nera', 'Escort the moon-silver shipment through the marsh road.', 1200, 0, '', 'partially_paid', '', '2026-02-01', '2026-02-10', NULL, '2026-02-01 07:30:00', '2026-02-15 12:45:00'),
	('quest_lost_heirloom', 'default', 'The Lost Heirloom', 'Innkeeper Bram', 'Recover the silver locket stolen by the river bandits.', 500, 0, '', 'completed', 'Bandits fled downstream. Locket was in the camp.', '2026-02-08', '2026-02-12', NULL, '2026-02-08 14:00:00', '2026-02-12 17:00:00');

-- ── Loot items ───────────────────────────────────────────────────────────

INSERT INTO loot_items (id, campaign_id, name, source, status, quantity, holder, notes, created_at, updated_at)
VALUES
	('loot_wyvern_necklace', 'default', 'Wyvern Tooth Necklace', 'Clear the Old Watchtower', 'held', 1, 'Mira', 'Still wrapped in oily cloth.', '2026-03-02 13:00:00', '2026-03-02 13:00:00'),
	('loot_emerald_idol', 'default', 'Emerald Idol', 'Moonlit Escort', 'recognized', 1, 'Quartermaster', 'Recognized after guild appraisal.', '2026-02-18 13:30:00', '2026-02-18 14:00:00'),
	('loot_ruby_crown', 'default', 'Cracked Ruby Crown', 'Moonlit Escort', 'sold', 1, 'Quartermaster', 'Sold below appraisal to move it quickly.', '2026-02-19 08:45:00', '2026-02-20 16:30:00'),
	('loot_silver_locket', 'default', 'Silver Locket', 'The Lost Heirloom', 'held', 1, 'Ragnar', 'Returned to Innkeeper Bram for the reward.', '2026-02-12 17:30:00', '2026-02-12 17:30:00'),
	('loot_bandit_map', 'default', 'Bandit Camp Map', 'River Bandits', 'held', 1, 'Mira', 'Shows supply routes along the Frostfall river.', '2026-02-12 17:45:00', '2026-02-12 17:45:00');

INSERT INTO loot_appraisals (id, loot_item_id, appraised_value, appraiser, notes, appraised_at, recognized_entry_id, created_at)
VALUES
	('appr_wyvern', 'loot_wyvern_necklace', 650, 'Guild Assayer', 'Teeth are chipped but still saleable.', '2026-03-02', NULL, '2026-03-02 13:15:00'),
	('appr_emerald', 'loot_emerald_idol', 800, 'Guild Assayer', 'Stonework matches the marsh shrine style.', '2026-02-18', 'je09', '2026-02-18 13:45:00'),
	('appr_ruby', 'loot_ruby_crown', 600, 'Guild Assayer', 'Several settings are bent and one ruby is missing.', '2026-02-19', 'je10', '2026-02-19 09:00:00');

-- ── Assets ───────────────────────────────────────────────────────────────

INSERT INTO loot_items (id, campaign_id, name, source, status, item_type, quantity, holder, notes, created_at, updated_at)
VALUES
	('asset_bag_of_holding', 'default', 'Bag of Holding', 'Party capital', 'held', 'asset', 1, 'Ragnar', 'Bought in Thornfield before departure. Holds the shared supply cache.', '2026-02-01 10:00:00', '2026-02-01 10:00:00'),
	('asset_spyglass', 'default', 'Dwarven Spyglass', 'Moonlit Escort', 'held', 'asset', 1, 'Mira', 'Recovered from the bandit lookout post. Excellent optics.', '2026-02-12 18:00:00', '2026-02-12 18:00:00');

-- ── Codex entries ────────────────────────────────────────────────────────

INSERT INTO codex_entries (id, campaign_id, type_id, name, title, location, faction, disposition, description, notes, created_at, updated_at)
VALUES
	('codex_elra', 'default', 'npc', 'Mayor Elra', 'Mayor of Thornfield', 'Thornfield', 'Town Council', 'friendly', 'Cautious halfling mayor who has governed Thornfield for twelve years.', 'Offered the watchtower bounty after goblin raids resumed.', '2026-02-25 09:00:00', '2026-02-25 09:00:00'),
	('codex_nera', 'default', 'npc', 'Guild Factor Nera', 'Merchant Guild Factor', 'Frostfall Keep', 'Merchant Guild', 'neutral', 'Sharp-eyed tiefling who brokers guild contracts from the keep.', 'Still owes 5 GP on the Moonlit Escort contract.', '2026-02-01 07:30:00', '2026-02-15 12:45:00'),
	('codex_pell', 'default', 'npc', 'Archivist Pell', 'Royal Archivist', 'Frostfall Keep', '', 'friendly', 'Elderly human scholar obsessed with pre-Sundering maps.', 'Nervous traveler. Insists on no fire near the scrolls.', '2026-03-01 08:00:00', '2026-03-01 08:00:00'),
	('codex_bram', 'default', 'npc', 'Innkeeper Bram', 'Proprietor, Silver Oak Inn', 'Thornfield', '', 'friendly', 'Broad-shouldered human who runs the only inn in town.', 'Grateful for the recovered locket. Gave a discount on rooms.', '2026-02-08 14:00:00', '2026-02-12 17:00:00'),
	('codex_ragnar', 'default', 'player', 'Ragnar', '', '', '', '', 'Half-orc fighter and party quartermaster.', 'Carries the Bag of Holding and manages supplies.', '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('codex_mira', 'default', 'player', 'Mira', '', '', '', '', 'Wood elf ranger and scout.', 'Expert tracker. Keeps the Dwarven Spyglass.', '2026-02-01 07:00:00', '2026-02-01 07:00:00');

INSERT INTO entity_references (id, campaign_id, source_type, source_id, source_name, target_type, target_name, created_at)
VALUES
	('ref_elra_wt', 'default', 'codex', 'codex_elra', 'Mayor Elra', 'quest', 'Clear the Old Watchtower', '2026-02-25 09:00:00'),
	('ref_nera_ml', 'default', 'codex', 'codex_nera', 'Guild Factor Nera', 'quest', 'Moonlit Escort', '2026-02-01 07:30:00'),
	('ref_pell_ea', 'default', 'codex', 'codex_pell', 'Archivist Pell', 'quest', 'Escort the Archivist', '2026-03-01 08:00:00'),
	('ref_bram_lh', 'default', 'codex', 'codex_bram', 'Innkeeper Bram', 'quest', 'The Lost Heirloom', '2026-02-08 14:00:00');

-- ── Notes ────────────────────────────────────────────────────────────────

INSERT INTO notes (id, campaign_id, title, body, created_at, updated_at)
VALUES
	('note_s1', 'default', 'Session 1: The Frostfall Campaign',
	 'The party pooled **150 GP** starting capital and set out from Thornfield. First stop: the Silver Oak Inn, where Ragnar broke the chandelier arm-wrestling a dwarf.

Expenses:
- Chandelier repair: 3 GP 5 SP
- Arrows and chalk: 1 GP 7 SP 5 CP
- Two nights lodging: 2 GP

Innkeeper Bram was surprisingly understanding about the chandelier.',
	 '2026-02-01 20:00:00', '2026-02-01 20:00:00'),

	('note_moonlit', 'default', 'Moonlit Escort Debrief',
	 'Moon-silver delivered intact. Ambush at the marsh crossing cost two healing potions. Guild Factor Nera paid **7 GP** up front, still owes **5 GP**.

The **Emerald Idol** was found in the bandit camp — recognized at 8 GP after guild appraisal. The Cracked Ruby Crown was also found but sold quickly at 4 GP (2 GP loss vs appraisal).

Lessons learned:
1. Always scout the marsh crossing before dawn
2. Healing potions are expensive — budget 2 GP per trip
3. Mira''s spyglass paid for itself on the lookout check',
	 '2026-02-15 20:00:00', '2026-02-15 20:00:00'),

	('note_wt_intel', 'default', 'Watchtower Recon Notes',
	 'Scouts report at least a dozen goblins and one **bugbear chief**. East wall has a collapsed section that could be used for entry.

Mayor Elra warned the bell may be **trapped** — the goblins rigged it with an alarm. Plan:
1. Mira approaches from the east breach at dusk
2. Ragnar draws attention at the main gate
3. Disable the bell trap before engaging the chief

Estimated cost: 2 GP for healing supplies, 1 GP for smoke bombs.',
	 '2026-02-28 14:00:00', '2026-02-28 14:00:00'),

	('note_s3', 'default', 'Session 3: The Archivist''s Road',
	 'Archivist Pell hired the party for **15 GP** (3 GP advance) to escort him to Frostfall Keep. He carries a locked case of pre-Sundering maps and insists on no fire within 10 feet of the scrolls.

The road passes through Thornfield and along the river where the Lost Heirloom bandits operated. Mira spotted old campfire remains — the bandits may have moved upstream.

Note: Pell mentioned a **second expedition** to the Sundering archives if this trip goes well. Could be worth 30+ GP.',
	 '2026-03-01 20:00:00', '2026-03-01 20:00:00');

INSERT INTO entity_references (id, campaign_id, source_type, source_id, source_name, target_type, target_name, created_at)
VALUES
	('ref_n_s1_ragnar', 'default', 'note', 'note_s1', 'Session 1: The Frostfall Campaign', 'person', 'Ragnar', '2026-02-01 20:00:00'),
	('ref_n_s1_bram', 'default', 'note', 'note_s1', 'Session 1: The Frostfall Campaign', 'person', 'Innkeeper Bram', '2026-02-01 20:00:00'),
	('ref_n_ml_nera', 'default', 'note', 'note_moonlit', 'Moonlit Escort Debrief', 'quest', 'Moonlit Escort', '2026-02-15 20:00:00'),
	('ref_n_ml_mira', 'default', 'note', 'note_moonlit', 'Moonlit Escort Debrief', 'person', 'Mira', '2026-02-15 20:00:00'),
	('ref_n_wt_elra', 'default', 'note', 'note_wt_intel', 'Watchtower Recon Notes', 'quest', 'Clear the Old Watchtower', '2026-02-28 14:00:00'),
	('ref_n_wt_mira', 'default', 'note', 'note_wt_intel', 'Watchtower Recon Notes', 'person', 'Mira', '2026-02-28 14:00:00'),
	('ref_n_wt_ragnar', 'default', 'note', 'note_wt_intel', 'Watchtower Recon Notes', 'person', 'Ragnar', '2026-02-28 14:00:00'),
	('ref_n_s3_pell', 'default', 'note', 'note_s3', 'Session 3: The Archivist''s Road', 'person', 'Archivist Pell', '2026-03-01 20:00:00'),
	('ref_n_s3_mira', 'default', 'note', 'note_s3', 'Session 3: The Archivist''s Road', 'person', 'Mira', '2026-03-01 20:00:00'),
	('ref_n_s3_lh', 'default', 'note', 'note_s3', 'Session 3: The Archivist''s Road', 'quest', 'The Lost Heirloom', '2026-03-01 20:00:00');

COMMIT;
