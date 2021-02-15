package errors

import (
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/pkg/exitcode"
)

var (
	ErrNotEnoughArgs           = errors.New("not enough args")
	ErrInvalidActionParamsType = errors.New("invalid action params type")
	ErrInvalidVersionFormat    = errors.New("invalid version format")
)

type GoMigrateError struct {
	Err      error
	ExitCode exitcode.ExitCode
}

func (e *GoMigrateError) Error() string {
	if e.Err == nil {
		return ""
	}

	return e.Err.Error()
}

func ErrorExitCode(err error) exitcode.ExitCode {
	if err == nil {
		return exitcode.ExitCodeOK
	}

	var goMigrateErr *GoMigrateError
	if errors.As(err, &goMigrateErr) {
		if goMigrateErr.ExitCode != exitcode.ExitCodeOK {
			return goMigrateErr.ExitCode
		}

		if goMigrateErr.Err != nil {
			return ErrorExitCode(goMigrateErr.Err)
		}
	}

	return exitcode.ExitCodeUnspecified
}
