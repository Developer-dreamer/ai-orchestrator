-- +goose Up
-- +goose StatementBegin
ALTER TABLE prompts ADD COLUMN response TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE prompts DROP COLUMN IF EXISTS response;
-- +goose StatementEnd
