-- +gomigrate Up
-- +gomigrate StatementBegin
ALTER TABLE some_table ALTER COLUMN password TYPE VARCHAR(80);
-- +gomigrate StatementEnd

-- +gomigrate Down
-- +gomigrate StatementBegin
ALTER TABLE some_table ALTER COLUMN password TYPE VARCHAR(50);
-- +gomigrate StatementEnd
