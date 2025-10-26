CREATE TABLE IF NOT EXISTS performance_logs (
    timestamp DateTime64(3) DEFAULT now64(),
    endpoint String,
    duration Int64,
    status_code Int32,
    cache_hit Bool,
    
    INDEX idx_endpoint endpoint TYPE bloom_filter GRANULARITY 1,
    INDEX idx_status_code status_code TYPE minmax GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, endpoint, status_code)
TTL timestamp + INTERVAL 30 DAY;