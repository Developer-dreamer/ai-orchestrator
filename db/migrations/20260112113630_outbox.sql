-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox
(
    id             UUID PRIMARY KEY,              -- Event ID
    aggregate_type VARCHAR(255) NOT NULL,         -- Entity type
    aggregate_id   UUID         NOT NULL,         -- Entity ID
    event_type     VARCHAR(255) NOT NULL,         -- Event Type (eg. "PromptCreated")
    payload        JSONB        NOT NULL,         -- Prompt struct JSON
    status         VARCHAR(50) DEFAULT 'pending', -- PENDING, PROCESSING, COMPLETED, FAILED
    retry_count    INT         DEFAULT 0,         -- Hw much tries for resending
    error_message  TEXT,                          -- Failed sending error log
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    processed_at   TIMESTAMPTZ                    -- When sent to broker
);

-- 2. Index for quick searching of unsent events
-- Critical for Relay-process, which will be reading data
CREATE INDEX idx_outbox_status_created_at ON outbox (status, created_at) WHERE status = 'pending';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_outbox_status_created_at;
DROP TABLE IF EXISTS outbox;
-- +goose StatementEnd