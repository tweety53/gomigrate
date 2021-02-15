package repo

type MigrationRepo interface {
	EnsureDBVersion() (string, error)
	GetDBVersion() (string, error)
	CreateVersionTable() error
	GetMigrationsHistory(limit int) (MigrationRecords, error)
	InsertVersion(v string) error
	InsertUnAppliedVersion(v string) error
	UpdateApplyTime(v string) error
	DeleteVersion(v string) error
	LockVersion(v string) error
}

type DBOperationRepo interface {
	TruncateDatabase() error
	GetForeignKeys(tableName string) (ForeignKeys, error)
	DropForeignKey(tableName string, fkName string) error
	DropTable(tableName string) error
	AllTableNames() ([]string, error)
}

type MigrationRecord struct {
	Version   string
	ApplyTime int
}

type MigrationRecords []*MigrationRecord
