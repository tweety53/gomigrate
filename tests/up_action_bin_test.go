// +build test_integration

package tests

import (
	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/sqldialect"
	"github.com/tweety53/gomigrate/pkg/config"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"testing"
)

func Test_UpAction(t *testing.T) {
	conf, err := config.BuildFromFile(`testdata/configs/bin_test_up_gomigrate.yaml`)
	if err != nil {
		log.Fatal(err)
	}

	db := getDb(conf)
	defer db.Close()
	dialect, err := sqldialect.InitDialect(conf.SQLDialect, conf.MigrationTable)
	if err != nil {
		log.Fatal(err)
	}

	mRepo := repo.NewMigrationsRepository(db, dialect)
	dboRepo := repo.NewDBOperationsRepository(db, dialect)
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantRecords repo.MigrationRecords
	}{
		{
			name: "up action: all",
			args: []string{},
			wantRecords: repo.MigrationRecords{
				&repo.MigrationRecord{
					Version: "m000000_000003_add_another_table",
				},
				&repo.MigrationRecord{
					Version: "m000000_000002_alter_some_table_column",
				},
				&repo.MigrationRecord{
					Version: "m000000_000001_add_some_table",
				},
			},
		},
		{
			name: "up action: one",
			args: []string{"1"},
			wantRecords: repo.MigrationRecords{
				&repo.MigrationRecord{
					Version: "m000000_000001_add_some_table",
				},
			},
		},
		{
			name: "up action: two",
			args: []string{"2"},
			wantRecords: repo.MigrationRecords{
				&repo.MigrationRecord{
					Version: "m000000_000002_alter_some_table_column",
				},
				&repo.MigrationRecord{
					Version: "m000000_000001_add_some_table",
				},
			},
		},
	}
	for _, test := range tests {
		err := dboRepo.TruncateDatabase()
		if err != nil {
			log.Fatal(err)
		}
		t.Run(test.name, func(t *testing.T) {
			args := []string{`-config`,
				`testdata/configs/bin_test_up_gomigrate.yaml`,
				`up`}

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

			var limit int
			if len(test.args) > 0 {
				limit, err = strconv.Atoi(test.args[0])
				if err != nil {
					log.Fatal(err)
				}
			}

			history, err := mRepo.GetMigrationsHistory(limit)
			require.Equal(t, len(history), len(test.wantRecords))
			for i := range history {
				require.Equal(t, test.wantRecords[i].Version, history[i].Version)
				require.Greater(t, history[i].ApplyTime, 0)
			}
		})
	}

	err = dboRepo.TruncateDatabase()
	if err != nil {
		log.Fatal(err)
	}
}

func Test_UpActionParallel(t *testing.T) {
	var (
		successCnt int64
		failCnt    int64
	)

	wantRecords := repo.MigrationRecords{
		&repo.MigrationRecord{
			Version: "m000000_000003_add_another_table",
		},
		&repo.MigrationRecord{
			Version: "m000000_000002_alter_some_table_column",
		},
		&repo.MigrationRecord{
			Version: "m000000_000001_add_some_table",
		},
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "up action: all runner 1",
			args: []string{},
		},
		{
			name: "up action: all runner 2",
			args: []string{},
		},
		{
			name: "up action: all runner 3",
			args: []string{},
		},
	}

	t.Run("wrapper for parallel runs waiting", func(t *testing.T) {
		for _, test := range tests {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				args := []string{`-config`,
					`testdata/configs/bin_test_up_gomigrate.yaml`,
					`up`}

				args = append(args, test.args...)

				cmd := exec.Command(
					binaryPath,
					args...)

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					atomic.AddInt64(&failCnt, 1)
				} else {
					atomic.AddInt64(&successCnt, 1)
				}
			})
		}
	})

	conf, err := config.BuildFromFile(`testdata/configs/bin_test_up_gomigrate.yaml`)
	if err != nil {
		log.Fatal(err)
	}

	db := getDb(conf)
	defer db.Close()
	dialect, err := sqldialect.InitDialect(conf.SQLDialect, conf.MigrationTable)
	if err != nil {
		log.Fatal(err)
	}

	mRepo := repo.NewMigrationsRepository(db, dialect)
	dboRepo := repo.NewDBOperationsRepository(db, dialect)

	history, err := mRepo.GetMigrationsHistory(0)
	require.Equal(t, len(wantRecords), len(history))
	for i := range history {
		require.Equal(t, wantRecords[i].Version, history[i].Version)
		require.Greater(t, history[i].ApplyTime, 0)
	}

	require.Equal(t, int64(1), atomic.LoadInt64(&successCnt))
	require.Equal(t, int64(2), atomic.LoadInt64(&failCnt))

	err = dboRepo.TruncateDatabase()
	if err != nil {
		log.Fatal(err)
	}
}
