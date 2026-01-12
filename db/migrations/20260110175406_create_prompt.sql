-- +goose Up
-- +goose StatementBegin
CREATE TABLE prompts
(
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    text       TEXT NOT NULL,
    status     VARCHAR(25) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

CREATE INDEX idx_prompts_user_id ON prompts (user_id);

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_prompts_user_id;
DROP TABLE prompts;
-- +goose StatementEnd
