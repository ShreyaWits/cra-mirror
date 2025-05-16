-- This is the initial schema migration file
-- It ensures that we have the required keyspace and table structure

-- Ensure the keyspace exists (will be ignored if already exists)
CREATE KEYSPACE IF NOT EXISTS protectedlink
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

-- Switch to our keyspace
USE protectedlink;

-- Create the main table if it doesn't exist
CREATE TABLE IF NOT EXISTS protectedLink (
    id UUID PRIMARY KEY,
    user_id TEXT,
    name TEXT,
    request_type TEXT,
    model_type TEXT,
    email TEXT,
    expire_in TEXT,
    otp_required BOOLEAN,
    phone TEXT,
    channel_type TEXT,
    data TEXT -- Storing JSON as text
);

-- Add any other tables or indices needed for the application 