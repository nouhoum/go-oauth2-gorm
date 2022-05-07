package gooauth2gorm

import "time"

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
	ConnMaxLifetime time.Duration
	DBType          DBType
}
