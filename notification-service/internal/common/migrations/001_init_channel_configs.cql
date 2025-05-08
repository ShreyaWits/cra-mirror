CREATE TABLE IF NOT EXISTS channel_configs (
    service TEXT PRIMARY KEY,
    primary_provider TEXT,
    fallback_provider TEXT
);
