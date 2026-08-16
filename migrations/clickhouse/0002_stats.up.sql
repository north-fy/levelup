CREATE TABLE IF NOT EXISTS quest_completed (
    user_id      UInt64,
    branch_id    UInt64,
    roadmap_id   UInt64,
    quest_id     UInt64,
    xp           Int32,
    gold         Int32,
    hours        Int32,
    completed_at DateTime
) ENGINE = MergeTree
PARTITION BY toYYYYMM(completed_at)
ORDER BY (user_id, completed_at);

CREATE TABLE IF NOT EXISTS purchase (
    user_id      UInt64,
    item_id      UInt64,
    price        Int32,
    purchased_at DateTime
) ENGINE = MergeTree
PARTITION BY toYYYYMM(purchased_at)
ORDER BY (user_id, purchased_at);