-- +gomigrate Up
-- +gomigrate StatementBegin
CREATE TABLE some_table
(
    user_id    serial PRIMARY KEY,
    username   VARCHAR(50) UNIQUE  NOT NULL,
    password   VARCHAR(50)         NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    created_on TIMESTAMP           NOT NULL,
    last_login TIMESTAMP
);
-- +gomigrate StatementEnd

-- +gomigrate Down
-- +gomigrate StatementBegin
DROP TABLE some_table;
-- +gomigrate StatementEnd
