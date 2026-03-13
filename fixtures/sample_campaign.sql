BEGIN TRANSACTION;

INSERT INTO campaigns (id, name, created_at, updated_at)
VALUES ('sample-campaign', 'Frostfall Campaign', '2026-02-01 07:00:00', '2026-02-01 07:00:00');

INSERT INTO settings (key, value) VALUES ('active_campaign_id', 'sample-campaign')
ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP;

INSERT INTO accounts (id, campaign_id, code, name, type, active, created_at, updated_at)
VALUES
	('sc_party_cash', 'sample-campaign', '1000', 'Party Cash', 'asset', 1, '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('sc_quest_receivable', 'sample-campaign', '1100', 'Quest Receivable', 'asset', 1, '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('sc_loot_inventory', 'sample-campaign', '1200', 'Loot Inventory', 'asset', 1, '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('sc_party_equity', 'sample-campaign', '3000', 'Party Equity', 'equity', 1, '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('sc_quest_income', 'sample-campaign', '4000', 'Quest Income', 'income', 1, '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('sc_unrealized_loot_gain', 'sample-campaign', '4200', 'Unrealized Loot Gain', 'income', 1, '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('sc_arrows_ammunition', 'sample-campaign', '5100', 'Arrows & Ammunition', 'expense', 1, '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('sc_tavern_reparations', 'sample-campaign', '5125', 'Tavern Reparations', 'expense', 1, '2026-02-01 09:00:00', '2026-02-01 09:00:00'),
	('sc_loss_on_sale', 'sample-campaign', '5400', 'Loss on Sale of Loot', 'expense', 1, '2026-02-01 07:00:00', '2026-02-01 07:00:00');

INSERT INTO journal_entries (id, campaign_id, entry_number, status, entry_date, description, reverses_entry_id, created_at, posted_at, reversed_at)
VALUES
	('je_opening_capital', 'sample-campaign', 1, 'posted', '2026-02-01', 'Party capital for the Frostfall campaign', NULL, '2026-02-01 09:00:00', '2026-02-01 09:00:00', NULL),
	('je_tavern_damage', 'sample-campaign', 2, 'posted', '2026-02-02', 'Pay for broken inn chandelier', NULL, '2026-02-02 10:15:00', '2026-02-02 10:15:00', NULL),
	('je_arrow_restock', 'sample-campaign', 3, 'posted', '2026-02-03', 'Restock arrows and chalk', NULL, '2026-02-03 08:30:00', '2026-02-03 08:30:00', NULL),
	('je_moonlit_earned', 'sample-campaign', 4, 'posted', '2026-02-10', 'Moonlit Escort earned', NULL, '2026-02-10 18:20:00', '2026-02-10 18:20:00', NULL),
	('je_moonlit_partial', 'sample-campaign', 5, 'posted', '2026-02-15', 'Guild factor sent the first pouch', NULL, '2026-02-15 12:45:00', '2026-02-15 12:45:00', NULL),
	('je_emerald_recognition', 'sample-campaign', 6, 'posted', '2026-02-18', 'Recognize appraisal: Emerald Idol', NULL, '2026-02-18 14:00:00', '2026-02-18 14:00:00', NULL),
	('je_ruby_recognition', 'sample-campaign', 7, 'posted', '2026-02-19', 'Recognize appraisal: Cracked Ruby Crown', NULL, '2026-02-19 09:10:00', '2026-02-19 09:10:00', NULL),
	('je_ruby_sale', 'sample-campaign', 8, 'posted', '2026-02-20', 'Sold cracked ruby crown below appraisal', NULL, '2026-02-20 16:30:00', '2026-02-20 16:30:00', NULL);

INSERT INTO journal_lines (id, journal_entry_id, line_number, account_id, memo, debit_amount, credit_amount)
VALUES
	('jl_opening_capital_1', 'je_opening_capital', 1, (SELECT id FROM accounts WHERE code = '1000' AND campaign_id = 'sample-campaign'), 'Initial purse and guild float', 15000, 0),
	('jl_opening_capital_2', 'je_opening_capital', 2, (SELECT id FROM accounts WHERE code = '3000' AND campaign_id = 'sample-campaign'), 'Initial purse and guild float', 0, 15000),
	('jl_tavern_damage_1', 'je_tavern_damage', 1, (SELECT id FROM accounts WHERE code = '5125' AND campaign_id = 'sample-campaign'), 'Silver Oak Inn settlement', 350, 0),
	('jl_tavern_damage_2', 'je_tavern_damage', 2, (SELECT id FROM accounts WHERE code = '1000' AND campaign_id = 'sample-campaign'), 'Silver Oak Inn settlement', 0, 350),
	('jl_arrow_restock_1', 'je_arrow_restock', 1, (SELECT id FROM accounts WHERE code = '5100' AND campaign_id = 'sample-campaign'), 'Arrows and chalk', 175, 0),
	('jl_arrow_restock_2', 'je_arrow_restock', 2, (SELECT id FROM accounts WHERE code = '1000' AND campaign_id = 'sample-campaign'), 'Arrows and chalk', 0, 175),
	('jl_moonlit_earned_1', 'je_moonlit_earned', 1, (SELECT id FROM accounts WHERE code = '1100' AND campaign_id = 'sample-campaign'), 'Quest completion: Moonlit Escort', 1200, 0),
	('jl_moonlit_earned_2', 'je_moonlit_earned', 2, (SELECT id FROM accounts WHERE code = '4000' AND campaign_id = 'sample-campaign'), 'Quest completion: Moonlit Escort', 0, 1200),
	('jl_moonlit_partial_1', 'je_moonlit_partial', 1, (SELECT id FROM accounts WHERE code = '1000' AND campaign_id = 'sample-campaign'), 'Quest payment: Moonlit Escort', 700, 0),
	('jl_moonlit_partial_2', 'je_moonlit_partial', 2, (SELECT id FROM accounts WHERE code = '1100' AND campaign_id = 'sample-campaign'), 'Quest payment: Moonlit Escort', 0, 700),
	('jl_emerald_recognition_1', 'je_emerald_recognition', 1, (SELECT id FROM accounts WHERE code = '1200' AND campaign_id = 'sample-campaign'), 'Recognized appraisal: Emerald Idol', 800, 0),
	('jl_emerald_recognition_2', 'je_emerald_recognition', 2, (SELECT id FROM accounts WHERE code = '4200' AND campaign_id = 'sample-campaign'), 'Recognized appraisal: Emerald Idol', 0, 800),
	('jl_ruby_recognition_1', 'je_ruby_recognition', 1, (SELECT id FROM accounts WHERE code = '1200' AND campaign_id = 'sample-campaign'), 'Recognized appraisal: Cracked Ruby Crown', 600, 0),
	('jl_ruby_recognition_2', 'je_ruby_recognition', 2, (SELECT id FROM accounts WHERE code = '4200' AND campaign_id = 'sample-campaign'), 'Recognized appraisal: Cracked Ruby Crown', 0, 600),
	('jl_ruby_sale_1', 'je_ruby_sale', 1, (SELECT id FROM accounts WHERE code = '1000' AND campaign_id = 'sample-campaign'), 'Sale proceeds: Cracked Ruby Crown', 400, 0),
	('jl_ruby_sale_2', 'je_ruby_sale', 2, (SELECT id FROM accounts WHERE code = '5400' AND campaign_id = 'sample-campaign'), 'Loss on sale: Cracked Ruby Crown', 200, 0),
	('jl_ruby_sale_3', 'je_ruby_sale', 3, (SELECT id FROM accounts WHERE code = '1200' AND campaign_id = 'sample-campaign'), 'Remove inventory: Cracked Ruby Crown', 0, 600);

INSERT INTO quests (
	id, campaign_id, title, patron, description, promised_base_reward, partial_advance, bonus_conditions,
	status, notes, accepted_on, completed_on, closed_on, created_at, updated_at
)
VALUES
	('quest_clear_watchtower', 'sample-campaign', 'Clear the Old Watchtower', 'Mayor Elra', 'Drive out the goblin squatters and recover the town bell.', 2000, 0, 'Bonus 5 GP if the bell is recovered intact', 'offered', '', NULL, NULL, NULL, '2026-02-25 09:00:00', '2026-02-25 09:00:00'),
	('quest_archivist_escort', 'sample-campaign', 'Escort the Archivist', 'Archivist Pell', 'Protect the archivist on the road to Frostfall Keep.', 1500, 300, 'Extra 2 GP if no pages are singed', 'accepted', '', '2026-03-01', NULL, NULL, '2026-03-01 08:00:00', '2026-03-01 08:00:00'),
	('quest_moonlit_escort', 'sample-campaign', 'Moonlit Escort', 'Guild Factor Nera', 'Escort the moon-silver shipment through the marsh road.', 1200, 0, '', 'partially_paid', '', '2026-02-01', '2026-02-10', NULL, '2026-02-01 07:30:00', '2026-02-15 12:45:00');

INSERT INTO loot_items (id, campaign_id, name, source, status, quantity, holder, notes, created_at, updated_at)
VALUES
	('loot_wyvern_necklace', 'sample-campaign', 'Wyvern Tooth Necklace', 'Clear the Old Watchtower', 'held', 1, 'Mira', 'Still wrapped in oily cloth.', '2026-03-02 13:00:00', '2026-03-02 13:00:00'),
	('loot_emerald_idol', 'sample-campaign', 'Emerald Idol', 'Moonlit Escort', 'recognized', 1, 'Quartermaster', 'Recognized after guild appraisal.', '2026-02-18 13:30:00', '2026-02-18 14:00:00'),
	('loot_ruby_crown', 'sample-campaign', 'Cracked Ruby Crown', 'Moonlit Escort', 'sold', 1, 'Quartermaster', 'Sold below appraisal to move it quickly.', '2026-02-19 08:45:00', '2026-02-20 16:30:00');

INSERT INTO loot_appraisals (id, loot_item_id, appraised_value, appraiser, notes, appraised_at, recognized_entry_id, created_at)
VALUES
	('appraisal_wyvern_necklace', 'loot_wyvern_necklace', 650, 'Guild Assayer', 'Teeth are chipped but still saleable.', '2026-03-02', NULL, '2026-03-02 13:15:00'),
	('appraisal_emerald_idol', 'loot_emerald_idol', 800, 'Guild Assayer', 'Stonework matches the marsh shrine style.', '2026-02-18', 'je_emerald_recognition', '2026-02-18 13:45:00'),
	('appraisal_ruby_crown', 'loot_ruby_crown', 600, 'Guild Assayer', 'Several settings are bent and one ruby is missing.', '2026-02-19', 'je_ruby_recognition', '2026-02-19 09:00:00');

INSERT INTO codex_entries (id, campaign_id, type_id, name, title, location, faction, disposition, description, notes, created_at, updated_at)
VALUES
	('codex_mayor_elra', 'sample-campaign', 'npc', 'Mayor Elra', 'Mayor of Thornfield', 'Thornfield', 'Town Council', 'friendly', 'Cautious halfling mayor who has governed Thornfield for twelve years.', 'Offered the watchtower bounty after goblin raids resumed.', '2026-02-25 09:00:00', '2026-02-25 09:00:00'),
	('codex_guild_factor_nera', 'sample-campaign', 'npc', 'Guild Factor Nera', 'Merchant Guild Factor', 'Frostfall Keep', 'Merchant Guild', 'neutral', 'Sharp-eyed tiefling who brokers guild contracts from the keep.', 'Still owes 5 GP on the Moonlit Escort contract.', '2026-02-01 07:30:00', '2026-02-15 12:45:00'),
	('codex_archivist_pell', 'sample-campaign', 'npc', 'Archivist Pell', 'Royal Archivist', 'Frostfall Keep', '', 'friendly', 'Elderly human scholar obsessed with pre-Sundering maps.', 'Nervous traveler. Insists on no fire near the scrolls.', '2026-03-01 08:00:00', '2026-03-01 08:00:00'),
	('codex_ragnar', 'sample-campaign', 'player', 'Ragnar', '', '', '', '', 'Half-orc fighter and party quartermaster.', '', '2026-02-01 07:00:00', '2026-02-01 07:00:00'),
	('codex_mira', 'sample-campaign', 'player', 'Mira', '', '', '', '', 'Wood elf ranger and scout.', '', '2026-02-01 07:00:00', '2026-02-01 07:00:00');

INSERT INTO entity_references (id, campaign_id, source_type, source_id, source_name, target_type, target_name, created_at)
VALUES
	('ref_elra_watchtower', 'sample-campaign', 'codex', 'codex_mayor_elra', 'Mayor Elra', 'quest', 'Clear the Old Watchtower', '2026-02-25 09:00:00'),
	('ref_nera_moonlit', 'sample-campaign', 'codex', 'codex_guild_factor_nera', 'Guild Factor Nera', 'quest', 'Moonlit Escort', '2026-02-01 07:30:00'),
	('ref_pell_escort', 'sample-campaign', 'codex', 'codex_archivist_pell', 'Archivist Pell', 'quest', 'Escort the Archivist', '2026-03-01 08:00:00');

INSERT INTO notes (id, campaign_id, title, body, created_at, updated_at)
VALUES
	('note_session_1', 'sample-campaign', 'Session 1: The Frostfall Campaign', 'The party pooled 150 GP starting capital and set out from Thornfield. First stop: the Silver Oak Inn, where Ragnar broke the chandelier arm-wrestling a dwarf.', '2026-02-01 20:00:00', '2026-02-01 20:00:00'),
	('note_moonlit_debrief', 'sample-campaign', 'Moonlit Escort Debrief', 'Moon-silver delivered intact. Ambush at the marsh crossing cost two healing potions. Guild Factor Nera paid 7 GP up front, still owes 5 GP. The emerald idol was found in the bandit camp.', '2026-02-15 20:00:00', '2026-02-15 20:00:00'),
	('note_watchtower_intel', 'sample-campaign', 'Watchtower Recon Notes', 'Scouts report at least a dozen goblins and one bugbear chief. East wall has a collapsed section that could be used for entry. Mayor Elra warned the bell may be trapped.', '2026-02-28 14:00:00', '2026-02-28 14:00:00');

INSERT INTO entity_references (id, campaign_id, source_type, source_id, source_name, target_type, target_name, created_at)
VALUES
	('ref_note_session1_ragnar', 'sample-campaign', 'note', 'note_session_1', 'Session 1: The Frostfall Campaign', 'person', 'Ragnar', '2026-02-01 20:00:00'),
	('ref_note_moonlit_nera', 'sample-campaign', 'note', 'note_moonlit_debrief', 'Moonlit Escort Debrief', 'quest', 'Moonlit Escort', '2026-02-15 20:00:00'),
	('ref_note_watchtower_elra', 'sample-campaign', 'note', 'note_watchtower_intel', 'Watchtower Recon Notes', 'quest', 'Clear the Old Watchtower', '2026-02-28 14:00:00');

COMMIT;
