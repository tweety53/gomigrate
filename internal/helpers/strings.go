package helpers

import "regexp"

var versionRegex = regexp.MustCompile("^(m(\\d{6}_?\\d{6})([\\w\\\\]+$))")

func ValidMigrationVersion(v string) bool {
	return versionRegex.MatchString(v)
}
