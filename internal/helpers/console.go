package helpers

import (
	"fmt"
	"strings"

	internalLog "github.com/tweety53/gomigrate/internal/log"
)

const LimitAll = "all"

func AskForConfirmation(text string) bool {
	internalLog.Warnln(text)

	var response string

	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}

	return processResponse(response)
}

func processResponse(response string) bool {
	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return false
	}
}

func ChooseLogText(n int, beforeRun bool) string {
	if n == 1 {
		if beforeRun {
			return "migration"
		}

		return "migration was"
	}

	if beforeRun {
		return "migrations"
	}

	return "migrations were"
}
