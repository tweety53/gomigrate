package helpers

import (
	"fmt"
	internalLog "github.com/tweety53/gomigrate/internal/log"
	"strings"
)

func AskForConfirmation(text string) bool {
	internalLog.Warnln(text)
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
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
