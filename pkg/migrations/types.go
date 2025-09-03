package migrations

type DBDriver string

const (
	DBDriverPostgres DBDriver = "postgres"
	DBDriverMySQL    DBDriver = "mysql"
)

// Valid checks if the DBDriver is valid.
func (d DBDriver) Valid() bool {
	switch d {
	case DBDriverPostgres, DBDriverMySQL:
		return true
	}
	return false
}
