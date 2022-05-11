package gooauth2gorm_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-oauth2/oauth2/v4/models"
	gooauth2gorm "github.com/nouhoum/go-oauth2-gorm"
	"github.com/stretchr/testify/assert"
)

func TestTokenStore(t *testing.T) {
	store := gooauth2gorm.NewTokenStoreWithDB(&gooauth2gorm.Config{DBType: gooauth2gorm.PostgresSQL}, db, "", 0)

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
	err := store.Create(ctx, info)
	assert.Nil(t, err, "store.Create")

	token, err := store.GetByCode(ctx, info.GetCode())
	assert.Nil(t, err, "store.GetByCode")
	assert.Equal(t, info.Code, token.GetCode())

	access, err := store.GetByAccess(ctx, info.GetAccess())
	assert.Nil(t, err)
	assert.Nil(t, access, "store.GetByAccess")

	//access token
	info2 := &models.Token{
		ClientID:      "client1",
		UserID:        "user1",
		RedirectURI:   "http://localhost",
		Scope:         "all",
		Access:        "access1",
		CodeCreateAt:  time.Now(),
		CodeExpiresIn: time.Second * 5,
	}

	err = store.Create(ctx, info2)
	assert.Nil(t, err, "store.Create")
	token, err = store.GetByAccess(ctx, info2.GetAccess())
	assert.Nil(t, err, "store.GetByAccess")
	assert.Equal(t, info2.Access, token.GetAccess())

	code, err := store.GetByCode(ctx, "...")
	assert.Nil(t, err)
	assert.Nil(t, code, "store.GetByCode")
}
