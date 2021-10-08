-- +goose Up
-- +goose StatementBegin

CREATE TABLE Chat(
    chat_id integer PRIMARY KEY,
    user_id integer NOT NULL,
    storage_name varchar(100) NOT NULL
);

CREATE INDEX chat_storage_name ON Chat(storage_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX chat_storage_name;
DROP TABLE Chat;

-- +goose StatementEnd
