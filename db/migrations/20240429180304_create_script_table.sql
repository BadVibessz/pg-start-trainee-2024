-- +goose Up
-- +goose StatementBegin
CREATE TABLE script
(
    id         bigserial primary key not null,
    command    text                  not null,
    output     text                  null,
    is_running boolean               not null,
    pid        bigint                null,
    created_at timestamp             not null default now(),
    updated_at timestamp             not null default now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE script;
-- +goose StatementEnd
