-- 1. Drop Indexes first (Optional, as dropping tables removes them, 
-- but good practice for explicit cleanup)
DROP INDEX IF EXISTS idx_schedule_entry_blocks_schedule_entry_id_entry_subgroup_id ON schedule_entry_blocks(schedule_entry_id, entry_subgroup_id);
DROP INDEX IF EXISTS idx_schedule_entry_blocks_schedule_entry_id ON schedule_entry_blocks(schedule_entry_id);
DROP INDEX IF EXISTS idx_cards_card_name ON cards(card_name);
DROP INDEX IF EXISTS idx_blocks_block_name ON blocks(block_name);
DROP INDEX IF EXISTS idx_schedule_entries_venue_id ON schedule_entries(venue_id);
DROP INDEX IF EXISTS idx_schedule_entries_conductor_id ON schedule_entries(conductor_id);
DROP INDEX IF EXISTS idx_schedule_entries_date;
DROP INDEX IF EXISTS idx_venues_name;
DROP INDEX IF EXISTS idx_conductors_full_name_user_id;

-- 2. Drop Tables in reverse order of dependencies
-- schedule_entry_blocks depends on schedule_entries and blocks
DROP TABLE IF EXISTS schedule_entry_blocks;

-- schedule_entries depends on conductors and venues
DROP TABLE IF EXISTS schedule_entries;

-- conductors depends on users
DROP TABLE IF EXISTS conductors;

-- blocks depends on cards
DROP TABLE IF EXISTS blocks;

-- 3. Finally drop the base table (users) and any other tables (events, venues)
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS venues;
DROP TABLE IF EXISTS cards;

-- 4. Drop the extension (only if you want a completely clean slate)
DROP EXTENSION IF EXISTS "pgcrypto";