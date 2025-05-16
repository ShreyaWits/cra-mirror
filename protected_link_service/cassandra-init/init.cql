CREATE KEYSPACE IF NOT EXISTS protectedlink
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE protectedlink;




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
