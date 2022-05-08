package gooauth2gorm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type Client struct {
	ID        string `gorm:"primaryKey"`
	Secret    string
	Domain    string
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ClientStore struct {
	db     *gorm.DB
	table  string
	stdout io.Writer
}

func NewClientStore(cfg *Config, table string) *ClientStore {
	if cfg == nil {
		panic(errors.New("db config is null"))
	}

	var d gorm.Dialector

	switch cfg.DBType {
	case PostgresSQL:
		d = postgres.New(postgres.Config{
			DSN: cfg.DSN,
		})
	case MySQL:
		d = mysql.New(mysql.Config{
			DSN: cfg.DSN,
		})
	case SQLite:
		d = sqlite.Open(cfg.DSN)
	case SQLServer:
		d = sqlserver.Open(cfg.DSN)
	case Clickhouse:
		d = clickhouse.Open(cfg.DSN)
	}

	db, err := gorm.Open(d)
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return NewClientStoreWithDB(cfg, db, table)
}

func NewClientStoreWithDB(cfg *Config, db *gorm.DB, table string) *ClientStore {
	cs := &ClientStore{
		db:    db,
		table: defaultClientTable,
	}

	if table != "" {
		cs.table = table
	}

	if !db.Migrator().HasTable(cs.table) {
		if err := db.Table(cs.table).Migrator().CreateTable(&Client{}); err != nil {
			panic(err)
		}
	}
	return cs
}

// GetByID retrieves and returns client information by id
func (cs *ClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var client Client
	if err := cs.db.WithContext(ctx).Table(cs.table).
		Where("id = ?", id).
		Find(&client).Error; err != nil {
		return nil, err
	}

	return cs.toClientInfo(client.Data)
}

// Create create oauth2 client
func (cs *ClientStore) Create(ctx context.Context, info oauth2.ClientInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	client := Client{
		ID:     info.GetID(),
		Secret: info.GetSecret(),
		Domain: info.GetDomain(),
		Data:   string(data),
	}

	return cs.db.WithContext(ctx).Table(cs.table).Create(client).Error
}

// RemoveByID delete the client by id
func (ts *ClientStore) RemoveByID(ctx context.Context, id string) error {
	return ts.db.WithContext(ctx).Table(ts.table).
		Where("id = ?", id).
		Delete(&Client{}).Error
}

func (cs *ClientStore) toClientInfo(data string) (oauth2.ClientInfo, error) {
	var cm models.Client
	err := json.Unmarshal([]byte(data), &cm)
	return &cm, err
}
