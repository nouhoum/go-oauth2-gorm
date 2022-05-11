package gooauth2gorm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

const (
	defaultTokenTable  = "oauth2_token"
	defaultClientTable = "oauth2_client"
)

type Token struct {
	gorm.Model
	ExpiredAt int64
	Code      string
	Access    string
	Refresh   string
	Data      string
}

type TokenStore struct {
	db     *gorm.DB
	table  string
	stdout io.Writer
	ticker *time.Ticker
}

func NewTokenStore(cfg *Config, table string, gcInterval int) *TokenStore {
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

	return NewTokenStoreWithDB(cfg, db, table, gcInterval)
}

func NewTokenStoreWithDB(cfg *Config, db *gorm.DB, table string, gcInterval int) *TokenStore {
	ts := &TokenStore{
		db:    db,
		table: defaultTokenTable,
	}

	if table != "" {
		ts.table = table
	}

	interval := 600
	if gcInterval > 0 {
		interval = gcInterval
	}
	ts.ticker = time.NewTicker(time.Second * time.Duration(interval))

	if !db.Migrator().HasTable(ts.table) {
		if err := db.Table(ts.table).Migrator().CreateTable(&Token{}); err != nil {
			panic(err)
		}
	}
	go ts.gc()
	return ts
}

func (ts TokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	token := &Token{
		Data: string(data),
	}

	if code := info.GetCode(); code != "" {
		token.Code = code
		token.ExpiredAt = info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()).Unix()
	} else {
		token.Access = info.GetAccess()
		token.ExpiredAt = info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()).Unix()

		if refresh := info.GetRefresh(); refresh != "" {
			token.Refresh = info.GetRefresh()
			token.ExpiredAt = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()).Unix()
		}
	}

	return ts.db.WithContext(ctx).Table(ts.table).Create(token).Error
}

// delete the authorization code
func (ts *TokenStore) RemoveByCode(ctx context.Context, code string) error {
	return ts.db.WithContext(ctx).Table(ts.table).
		Where("code = ?", code).
		Update("code", "").Error
}

// use the access token to delete the token information
func (ts *TokenStore) RemoveByAccess(ctx context.Context, access string) error {
	return ts.db.WithContext(ctx).Table(ts.table).
		Where("access = ?", access).
		Update("access", "").Error
}

// use the refresh token to delete the token information
func (ts *TokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	return ts.db.WithContext(ctx).Table(ts.table).
		Where("refresh = ?", refresh).
		Update("refresh", "").Error
}

// use the authorization code for token information data
func (ts *TokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	var token Token
	if err := ts.db.WithContext(ctx).Table(ts.table).
		Where("code = ?", code).
		Find(&token).Error; err != nil {
		return nil, err
	}
	if token.ID == 0 {
		return nil, nil
	}

	return ts.toTokenInfo(token.Data), nil
}

// GetByAccess use the access token for token information data
func (ts *TokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	var token Token
	if err := ts.db.WithContext(ctx).Table(ts.table).
		Where("access = ?", access).
		Find(&token).Error; err != nil {
		return nil, err
	}
	if token.ID == 0 {
		return nil, nil
	}

	return ts.toTokenInfo(token.Data), nil
}

//GetByRefresh use the refresh token for token information data
func (ts *TokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	var token Token
	if err := ts.db.WithContext(ctx).Table(ts.table).
		Where("refresh = ?", refresh).
		Find(&token).Error; err != nil {
		return nil, err
	}
	if token.ID == 0 {
		return nil, nil
	}

	return ts.toTokenInfo(token.Data), nil
}

// SetStdout set error output
func (ts *TokenStore) SetStdout(stdout io.Writer) *TokenStore {
	ts.stdout = stdout
	return ts
}

// Close close the store
func (ts *TokenStore) Close() {
	ts.ticker.Stop()
	//_ = ts.db.DB().Close()
}

func (ts *TokenStore) toTokenInfo(data string) oauth2.TokenInfo {
	var t models.Token
	err := json.Unmarshal([]byte(data), &t)
	if err != nil {
		return nil
	}
	return &t
}

func (ts *TokenStore) gc() {
	for range ts.ticker.C {
		ts.cleanup()
	}
}

func (ts TokenStore) cleanup() {
	for range ts.ticker.C {
		now := time.Now().Unix()
		var count int64
		if err := ts.db.Table(ts.table).Where("expired_at <= ?", now).Or("code = ? and access = ? AND refresh = ?", "", "", "").Count(&count).Error; err != nil {
			ts.errorf("%s\n", err)
			return
		}
		if count > 0 {
			if err := ts.db.Table(ts.table).Where("expired_at <= ?", now).Or("code = ? and access = ? AND refresh = ?", "", "", "").Unscoped().Delete(&Token{}).Error; err != nil {
				ts.errorf("%s\n", err)
			}
		}
	}
}

func (ts *TokenStore) errorf(format string, args ...interface{}) {
	if ts.stdout != nil {
		buf := fmt.Sprintf(format, args...)
		_, _ = ts.stdout.Write([]byte(buf))
	}
}
