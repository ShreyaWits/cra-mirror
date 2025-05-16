-- This is the rollback migration
-- It drops the tables and keyspace created in the up migration

-- Use the keyspace
USE protectedlink;

-- Drop the main table
DROP TABLE IF EXISTS protectedLink;

-- We're not dropping the keyspace here as it might contain other tables
-- If a complete cleanup is needed, uncomment the following:
-- DROP KEYSPACE IF EXISTS protectedlink; 