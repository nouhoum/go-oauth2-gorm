package gooauth2gorm

type DBType int

const (
	PostgresSQL DBType = iota
	MySQL
	SQLite
	SQLServer
	Clickhouse
)

type Config struct {
	DSN             string
	Table           string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
	DBType          DBType
}
