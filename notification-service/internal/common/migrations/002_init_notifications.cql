CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    recipient map<TEXT, TEXT>,
    channel TEXT,
    fallback_service TEXT,
    meta map<TEXT, TEXT>,
    tags list<TEXT>,
    template_id TEXT,
    primary_retry_count INT,
    fallback_retry_count INT,
    total_execution_time INT,
    notification_status TEXT,
    error_message TEXT,
    primary_service TEXT,
    is_primary_failed BOOLEAN,
    is_fallback_failed BOOLEAN,
    is_moved_to_dlq BOOLEAN,
    tracking_id TEXT
);
