package migration

import (
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"io"
	"regexp"
	"strings"
	"sync"
)

var (
	matchSQLComments = regexp.MustCompile(`(?m)^--.*$[\r\n]*`)
	matchEmptyEOL    = regexp.MustCompile(`(?m)^$[\r\n]*`) // TODO: Duplicate
)

func clearStatement(s string) string {
	s = matchSQLComments.ReplaceAllString(s, ``)
	return matchEmptyEOL.ReplaceAllString(s, ``)
}

type parserState int

const (
	start                       parserState = iota // 0
	gomigrateUp                                    // 1
	gomigrateStatementBeginUp                      // 2
	gomigrateStatementEndUp                        // 3
	gomigrateDown                                  // 4
	gomigrateStatementBeginDown                    // 5
	gomigrateStatementEndDown                      // 6
)

type stateMachine parserState

func (s *stateMachine) Get() parserState {
	return parserState(*s)
}
func (s *stateMachine) Set(new parserState) {
	log.Debugf("StateMachine: %v => %v", *s, new)
	*s = stateMachine(new)
}

const scanBufSize = 4 * 1024 * 1024

var matchEmptyLines = regexp.MustCompile(`^\s*$`)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, scanBufSize)
	},
}

// Split given SQL script into individual statements and return
// SQL statements for given direction (up=true, down=false).
//
// The base case is to simply split on semicolons, as these
// naturally terminate a statement.
//
// However, more complex cases like pl/pgsql can have semicolons
// within a statement. For these cases, we provide the explicit annotations
// 'StatementBegin' and 'StatementEnd' to allow the script to
// tell us to ignore semicolons.
func parseSQLMigration(r io.Reader, direction MigrationDirection) (stmts []string, useTx bool, err error) {
	var buf bytes.Buffer
	scanBuf := bufferPool.Get().([]byte)
	defer bufferPool.Put(scanBuf)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(scanBuf, scanBufSize)

	stateMachine := stateMachine(start)
	useTx = true

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "--") {
			cmd := strings.TrimSpace(strings.TrimPrefix(line, "--"))

			switch cmd {
			case "+gomigrate Up":
				switch stateMachine.Get() {
				case start:
					stateMachine.Set(gomigrateUp)
				default:
					return nil, false, errors.Errorf("duplicate '-- +gomigrate Up' annotations; stateMachine=%v", stateMachine)
				}
				continue

			case "+gomigrate Down":
				switch stateMachine.Get() {
				case gomigrateUp, gomigrateStatementEndUp:
					stateMachine.Set(gomigrateDown)
				default:
					return nil, false, errors.Errorf("must start with '-- +gomigrate Up' annotation, stateMachine=%v", stateMachine)
				}
				continue

			case "+gomigrate StatementBegin":
				switch stateMachine.Get() {
				case gomigrateUp, gomigrateStatementEndUp:
					stateMachine.Set(gomigrateStatementBeginUp)
				case gomigrateDown, gomigrateStatementEndDown:
					stateMachine.Set(gomigrateStatementBeginDown)
				default:
					return nil, false, errors.Errorf("'-- +gomigrate StatementBegin' must be defined after '-- +gomigrate Up' or '-- +gomigrate Down' annotation, stateMachine=%v", stateMachine)
				}
				continue

			case "+gomigrate StatementEnd":
				switch stateMachine.Get() {
				case gomigrateStatementBeginUp:
					stateMachine.Set(gomigrateStatementEndUp)
				case gomigrateStatementBeginDown:
					stateMachine.Set(gomigrateStatementEndDown)
				default:
					return nil, false, errors.New("'-- +gomigrate StatementEnd' must be defined after '-- +gomigrate StatementBegin'")
				}

			case "+gomigrate NO TRANSACTION":
				useTx = false
				continue

			default:
				// Ignore comments.
				log.Debugf("StateMachine: ignore comment")
				continue
			}
		}

		// Ignore empty lines.
		if matchEmptyLines.MatchString(line) {
			log.Debugf("StateMachine: ignore empty line")
			continue
		}

		// Write SQL line to a buffer.
		if _, err := buf.WriteString(line + "\n"); err != nil {
			return nil, false, errors.Wrap(err, "failed to write to buf")
		}

		// Read SQL body one by line, if we're in the right direction.
		//
		// 1) basic query with semicolon; 2) psql statement
		//
		// Export statement once we hit end of statement.
		switch stateMachine.Get() {
		case gomigrateUp, gomigrateStatementBeginUp, gomigrateStatementEndUp:
			if direction == migrationDirectionDown {
				buf.Reset()
				log.Debugf("StateMachine: ignore down")
				continue
			}
		case gomigrateDown, gomigrateStatementBeginDown, gomigrateStatementEndDown:
			if direction == migrationDirectionUp {
				buf.Reset()
				log.Debugf("StateMachine: ignore up")
				continue
			}
		default:
			return nil, false, errors.Errorf("failed to parse migration: unexpected state %q on line %q", stateMachine, line)
		}

		switch stateMachine.Get() {
		case gomigrateUp:
			if endsWithSemicolon(line) {
				stmts = append(stmts, buf.String())
				buf.Reset()
				log.Debugf("StateMachine: store simple Up query")
			}
		case gomigrateDown:
			if endsWithSemicolon(line) {
				stmts = append(stmts, buf.String())
				buf.Reset()
				log.Debugf("StateMachine: store simple Down query")
			}
		case gomigrateStatementEndUp:
			stmts = append(stmts, buf.String())
			buf.Reset()
			log.Debugf("StateMachine: store Up statement")
			stateMachine.Set(gomigrateUp)
		case gomigrateStatementEndDown:
			stmts = append(stmts, buf.String())
			buf.Reset()
			log.Debugf("StateMachine: store Down statement")
			stateMachine.Set(gomigrateDown)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, false, errors.Wrap(err, "failed to scan migration")
	}
	// EOF

	switch stateMachine.Get() {
	case start:
		return nil, false, errors.New("failed to parse migration: must start with '-- +gomigrate Up' annotation")
	case gomigrateStatementBeginUp, gomigrateStatementBeginDown:
		return nil, false, errors.New("failed to parse migration: missing '-- +gomigrate StatementEnd' annotation")
	}

	if bufferRemaining := strings.TrimSpace(buf.String()); len(bufferRemaining) > 0 {
		return nil, false, errors.Errorf("failed to parse migration: state %q, direction: %v: unexpected unfinished SQL query: %q: missing semicolon?", stateMachine, direction, bufferRemaining)
	}

	return stmts, useTx, nil
}

// Checks the line to see if the line has a statement-ending semicolon
// or if the line contains a double-dash comment.
func endsWithSemicolon(line string) bool {
	scanBuf := bufferPool.Get().([]byte)
	defer bufferPool.Put(scanBuf)

	prev := ""
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Buffer(scanBuf, scanBufSize)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "--") {
			break
		}
		prev = word
	}

	return strings.HasSuffix(prev, ";")
}
