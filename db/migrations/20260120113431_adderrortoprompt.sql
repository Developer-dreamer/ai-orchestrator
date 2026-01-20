-- +goose Up
-- +goose StatementBegin
ALTER TABLE prompts
    ADD COLUMN IF NOT EXISTS model_id VARCHAR(128),
    ADD COLUMN IF NOT EXISTS error TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE prompts
    DROP COLUMN IF EXISTS model_id,
    DROP COLUMN IF EXISTS error;
-- +goose StatementEnd
