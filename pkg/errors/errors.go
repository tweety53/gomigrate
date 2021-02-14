package errors

import (
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/pkg/exit_code"
)

var (
	ErrNotEnoughArgs           = errors.New("not enough args")
	ErrInvalidActionParamsType = errors.New("invalid action params type")
	ErrInvalidVersionFormat    = errors.New("invalid version format")
)

type GoMigrateError struct {
	Err      error
	ExitCode exit_code.ExitCode
}

func (e *GoMigrateError) Error() string {
	if e.Err == nil {
		return ""
	}

	return e.Err.Error()
}

func ErrorExitCode(err error) exit_code.ExitCode {
	if err == nil {
		return exit_code.ExitCodeOK
	} else if e, ok := err.(*GoMigrateError); ok && e.ExitCode != exit_code.ExitCodeOK {
		return e.ExitCode
	} else if ok && e.Err != nil {
		return ErrorExitCode(e.Err)
	}

	return exit_code.ExitCodeUnspecified
}
