CREATE KEYSPACE IF NOT EXISTS notifications
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE notifications;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name TEXT,
    email TEXT
);

