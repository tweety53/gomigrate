package version

import (
	"fmt"
	"regexp"
	"time"

	"github.com/tweety53/gomigrate/internal/migration"
)

// Migrations version prefix format.
const versionPrefixFormat = "m060102_150405"

//nolint:gochecknoglobals
var versionRegex = regexp.MustCompile(`^(m(\d{6}_?\d{6})([\w\\]+$))`)

func ValidMigrationVersion(v string) bool {
	return versionRegex.MatchString(v)
}

func BuildVersion(name string, mType migration.Type) string {
	versionPrefix := time.Now().Format(versionPrefixFormat)
	version := fmt.Sprintf("%s_%s", versionPrefix, name)

	return fmt.Sprintf("%s.%s", version, mType)
}
