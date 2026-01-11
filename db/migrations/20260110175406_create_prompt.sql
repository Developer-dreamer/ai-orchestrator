-- +goose Up
-- +goose StatementBegin
CREATE TABLE prompts
(
    id         UUID PRIMARY KEY,
    user_id    UUID,
    text       TEXT,
    status     varchar(25),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE prompts
-- +goose StatementEnd
