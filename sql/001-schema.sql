-- CREATE DATABASE IF NOT EXISTS dm_change_log;
-- USE dm_change_log;

CREATE TABLE IF NOT EXISTS change_event (
    id BINARY(16) PRIMARY KEY,
    event_time BIGINT(11),
    event_object_id VARCHAR(255),
    event_object_type VARCHAR(511),
    effected_service VARCHAR(255),
    source_service VARCHAR(255),
    correlation_id VARCHAR(255),
    user VARCHAR(255),
    reason VARCHAR(255),
    comment VARCHAR(511),
    event_type ENUM('create', 'update', 'soft_delete', 'soft_restore', 'hard_delete'),
    before_object BLOB,
    after_object BLOB
    -- TODO: hashes / indices
);