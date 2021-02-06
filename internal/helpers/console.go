package helpers

import (
	"fmt"
	internalLog "github.com/tweety53/gomigrate/internal/log"
	"log"
	"strings"
)

func AskForConfirmation(text string, defaultResponse bool) bool {
	internalLog.Warnln(text)
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(response) {
	case "y", "yes", "Y", "YES":
		return true
	case "n", "no", "N", "NO":
		return false
	default:
		return defaultResponse
	}
}
