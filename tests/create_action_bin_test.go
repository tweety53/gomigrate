// +build test_integration

package tests

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/version"
)

const (
	createActionTestDir = "testdata/create_test/"
)

func Test_CreateAction(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "create action: name only",
			args: []string{"test_name"},
		},
		{
			name: "create action: type go",
			args: []string{"test_name_go", "go"},
		},
		{
			name: "create action: type sql",
			args: []string{"test_name_sql", "sql"},
		},
		{
			name:    "create action: unknown type",
			args:    []string{"test_name_kek", "kek"},
			wantErr: true,
		},
		{
			name:    "create action: bad name pattern",
			args:    []string{"test_name_kek+"},
			wantErr: true,
		},
		{
			name: "create action: type sql safe=true",
			args: []string{"test_name_sql_safe", "sql", "true"},
		},
		{
			name: "create action: type sql safe=false",
			args: []string{"test_name_sql_non_safe", "sql", "false"},
		},
		{
			name: "create action: type go safe=true",
			args: []string{"test_name_go_safe", "go", "true"},
		},
		{
			name: "create action: type go safe=false",
			args: []string{"test_name_go_non_safe", "go", "false"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := []string{`-config`,
				`testdata/configs/bin_test_create_gomigrate.yaml`,
				`create`}

			args = append(args, test.args...)

			cmd := exec.Command(
				binaryPath,
				args...)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if test.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			var mType migration.Type
			if len(test.args) == 1 {
				mType = migration.TypeGo
			} else {
				mType = migration.Type(test.args[1])
			}
			_, err = os.Stat(createActionTestDir + version.BuildVersion(test.args[0], mType))

			require.NoError(t, err)
		})
	}
}
