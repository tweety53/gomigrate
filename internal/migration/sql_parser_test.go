package migration

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestSemicolons(t *testing.T) {
	t.Parallel()

	type testData struct {
		line   string
		result bool
	}

	tests := []testData{
		{line: "END;", result: true},
		{line: "END; -- comment", result: true},
		{line: "END   ; -- comment", result: true},
		{line: "END -- comment", result: false},
		{line: "END -- comment ;", result: false},
		{line: "END \" ; \" -- comment", result: false},
	}

	for _, test := range tests {
		r := endsWithSemicolon(test.line)
		if r != test.result {
			t.Errorf("incorrect semicolon. got %v, want %v", r, test.result)
		}
	}
}

func TestSplitStatements(t *testing.T) {
	t.Parallel()

	type testData struct {
		sql  string
		up   int
		down int
	}

	tt := []testData{
		{sql: multilineSQL, up: 4, down: 1},
		{sql: emptySQL, up: 0, down: 0},
		{sql: emptySQL2, up: 0, down: 0},
		{sql: functxt, up: 2, down: 2},
		{sql: mysqlChangeDelimiter, up: 4, down: 0},
		{sql: copyFromStdin, up: 1, down: 0},
		{sql: plpgsqlSyntax, up: 2, down: 2},
		{sql: plpgsqlSyntaxMixedStatements, up: 2, down: 2},
	}

	for i, test := range tt {
		// up
		stmts, _, err := parseSQLMigration(strings.NewReader(test.sql), migrationDirectionUp)
		if err != nil {
			t.Error(errors.Wrapf(err, "tt[%v] unexpected error", i))
		}
		if len(stmts) != test.up {
			t.Errorf("tt[%v] incorrect number of up stmts. got %v (%+v), want %v", i, len(stmts), stmts, test.up)
		}

		// down
		stmts, _, err = parseSQLMigration(strings.NewReader(test.sql), migrationDirectionDown)
		if err != nil {
			t.Error(errors.Wrapf(err, "tt[%v] unexpected error", i))
		}
		if len(stmts) != test.down {
			t.Errorf("tt[%v] incorrect number of down stmts. got %v (%+v), want %v", i, len(stmts), stmts, test.down)
		}
	}
}

func TestUseTransactions(t *testing.T) {
	//todo: implement NO TRANSACTION parsing tests
}

func TestParsingErrors(t *testing.T) {
	tt := []string{
		statementBeginNoStatementEnd,
		unfinishedSQL,
		noUpDownAnnotations,
		multiUpDown,
		downFirst,
	}
	for i, sql := range tt {
		_, _, err := parseSQLMigration(strings.NewReader(sql), migrationDirectionUp)
		if err == nil {
			t.Errorf("expected error on tt[%v] %q", i, sql)
		}
	}
}

var multilineSQL = `-- +gomigrate Up
CREATE TABLE post (
		id int NOT NULL,
		title text,
		body text,
		PRIMARY KEY(id)
);                  -- 1st stmt

-- comment
SELECT 2;           -- 2nd stmt
SELECT 3; SELECT 3; -- 3rd stmt
SELECT 4;           -- 4th stmt

-- +gomigrate Down
-- comment
DROP TABLE post;    -- 1st stmt
`

var functxt = `-- +gomigrate Up
CREATE TABLE IF NOT EXISTS histories (
	id                BIGSERIAL  PRIMARY KEY,
	current_value     varchar(2000) NOT NULL,
	created_at      timestamp with time zone  NOT NULL
);

-- +gomigrate StatementBegin
CREATE OR REPLACE FUNCTION histories_partition_creation( DATE, DATE )
returns void AS $$
DECLARE
	create_query text;
BEGIN
	FOR create_query IN SELECT
			'CREATE TABLE IF NOT EXISTS histories_'
			|| TO_CHAR( d, 'YYYY_MM' )
			|| ' ( CHECK( created_at >= timestamp '''
			|| TO_CHAR( d, 'YYYY-MM-DD 00:00:00' )
			|| ''' AND created_at < timestamp '''
			|| TO_CHAR( d + INTERVAL '1 month', 'YYYY-MM-DD 00:00:00' )
			|| ''' ) ) inherits ( histories );'
		FROM generate_series( $1, $2, '1 month' ) AS d
	LOOP
		EXECUTE create_query;
	END LOOP;  -- LOOP END
END;         -- FUNCTION END
$$
language plpgsql;
-- +gomigrate StatementEnd

-- +gomigrate Down
drop function histories_partition_creation(DATE, DATE);
drop TABLE histories;
`

var multiUpDown = `-- +gomigrate Up
CREATE TABLE post (
		id int NOT NULL,
		title text,
		body text,
		PRIMARY KEY(id)
);

-- +gomigrate Down
DROP TABLE post;

-- +gomigrate Up
CREATE TABLE fancier_post (
		id int NOT NULL,
		title text,
		body text,
		created_on timestamp without time zone,
		PRIMARY KEY(id)
);
`

var downFirst = `-- +gomigrate Down
DROP TABLE fancier_post;
`

var statementBeginNoStatementEnd = `-- +gomigrate Up
CREATE TABLE IF NOT EXISTS histories (
  id                BIGSERIAL  PRIMARY KEY,
  current_value     varchar(2000) NOT NULL,
  created_at      timestamp with time zone  NOT NULL
);

-- +gomigrate StatementBegin
CREATE OR REPLACE FUNCTION histories_partition_creation( DATE, DATE )
returns void AS $$
DECLARE
  create_query text;
BEGIN
  FOR create_query IN SELECT
      'CREATE TABLE IF NOT EXISTS histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || ' ( CHECK( created_at >= timestamp '''
      || TO_CHAR( d, 'YYYY-MM-DD 00:00:00' )
      || ''' AND created_at < timestamp '''
      || TO_CHAR( d + INTERVAL '1 month', 'YYYY-MM-DD 00:00:00' )
      || ''' ) ) inherits ( histories );'
    FROM generate_series( $1, $2, '1 month' ) AS d
  LOOP
    EXECUTE create_query;
  END LOOP;  -- LOOP END
END;         -- FUNCTION END
$$
language plpgsql;

-- +gomigrate Down
drop function histories_partition_creation(DATE, DATE);
drop TABLE histories;
`

var unfinishedSQL = `
-- +gomigrate Up
ALTER TABLE post

-- +gomigrate Down
`

var emptySQL = `-- +gomigrate Up
-- This is just a comment`

var emptySQL2 = `

-- comment
-- +gomigrate Up

-- comment
-- +gomigrate Down

`

var noUpDownAnnotations = `
CREATE TABLE post (
    id int NOT NULL,
    title text,
    body text,
    PRIMARY KEY(id)
);
`

var mysqlChangeDelimiter = `
-- +gomigrate Up
-- +gomigrate StatementBegin
DELIMITER | 
-- +gomigrate StatementEnd

-- +gomigrate StatementBegin
CREATE FUNCTION my_func( str CHAR(255) ) RETURNS CHAR(255) DETERMINISTIC
BEGIN 
  RETURN "Dummy Body"; 
END | 
-- +gomigrate StatementEnd

-- +gomigrate StatementBegin
DELIMITER ; 
-- +gomigrate StatementEnd

select my_func("123") from dual;
-- +gomigrate Down
`

var copyFromStdin = `
-- +gomigrate Up
-- +gomigrate StatementBegin
COPY public.django_content_type (id, app_label, model) FROM stdin;
1	admin	logentry
2	auth	permission
3	auth	group
4	auth	user
5	contenttypes	contenttype
6	sessions	session
\.
-- +gomigrate StatementEnd
`

var plpgsqlSyntax = `
-- +gomigrate Up
-- +gomigrate StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';
-- +gomigrate StatementEnd
-- +gomigrate StatementBegin
CREATE TRIGGER update_properties_updated_at BEFORE UPDATE ON properties FOR EACH ROW EXECUTE PROCEDURE  update_updated_at_column();
-- +gomigrate StatementEnd

-- +gomigrate Down
-- +gomigrate StatementBegin
DROP TRIGGER update_properties_updated_at
-- +gomigrate StatementEnd
-- +gomigrate StatementBegin
DROP FUNCTION update_updated_at_column()
-- +gomigrate StatementEnd
`

var plpgsqlSyntaxMixedStatements = `
-- +gomigrate Up
-- +gomigrate StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';
-- +gomigrate StatementEnd

CREATE TRIGGER update_properties_updated_at
BEFORE UPDATE
ON properties 
FOR EACH ROW EXECUTE PROCEDURE  update_updated_at_column();

-- +gomigrate Down
DROP TRIGGER update_properties_updated_at;
DROP FUNCTION update_updated_at_column();
`
