package gooauth2gorm_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-oauth2/oauth2/v4/models"
	gooauth2gorm "github.com/nouhoum/go-oauth2-gorm"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTokenStore(t *testing.T) {
	dsn := fmt.Sprintf("host=%s user=test password=test dbname=test port=%d sslmode=disable", dbHost, dbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.Nil(t, err, "gorm.Open")
	store := gooauth2gorm.NewTokenStoreWithDB(&gooauth2gorm.Config{DSN: dsn, DBType: gooauth2gorm.PostgresSQL}, db, "", 0)

	ctx := context.Background()
	info := &models.Token{
		ClientID:      "client1",
		UserID:        "user1",
		RedirectURI:   "http://localhost",
		Scope:         "all",
		Code:          "code1",
		CodeCreateAt:  time.Now(),
		CodeExpiresIn: time.Second * 5,
	}
	err = store.Create(ctx, info)
	assert.Nil(t, err, "store.Create")

	tInfo, err := store.GetByCode(ctx, info.GetCode())
	assert.Nil(t, err, "store.Create")
	assert.Nil(t, err, "store.GetByCode")

	assert.Equal(t, info.Code, tInfo.GetCode())
}
