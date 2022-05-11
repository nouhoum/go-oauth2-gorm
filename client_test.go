package gooauth2gorm_test

import (
	"context"
	"testing"

	"github.com/go-oauth2/oauth2/v4/models"
	gooauth2gorm "github.com/nouhoum/go-oauth2-gorm"
	"github.com/stretchr/testify/assert"
)

func TestClientStore(t *testing.T) {
	store := gooauth2gorm.NewClientStoreWithDB(&gooauth2gorm.Config{DBType: gooauth2gorm.PostgresSQL}, db, "")

	ctx := context.Background()
	info := &models.Client{
		ID:     "client1",
		Secret: "secret",
		UserID: "user1",
		Domain: "http://localhost",
	}
	err := store.Create(ctx, info)
	assert.Nil(t, err, "store.Create")

	client, err := store.GetByID(ctx, info.ID)
	assert.Nil(t, err, "store.GetByID")
	assert.Equal(t, info.ID, client.GetID())

	/*
		access, err := store.GetByAccess(ctx, info.GetAccess())
		assert.Nil(t, err)
		assert.Nil(t, access, "store.GetByAccess")

		//Remove by code
		err = store.RemoveByCode(ctx, info.GetCode())
		assert.Nil(t, err)
		unknownToken, err := store.GetByCode(ctx, info.GetCode())
		assert.Nil(t, err, "store.GetByCode")
		assert.Nil(t, unknownToken, "store.GetByCode")

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
		assert.NotNil(t, token.GetAccess())
		assert.Equal(t, info2.Access, token.GetAccess())

		code, err := store.GetByCode(ctx, "...")
		assert.Nil(t, err)
		assert.Nil(t, code, "store.GetByCode")

		//Remove access code
		err = store.RemoveByAccess(ctx, info2.GetAccess())
		//assert.Nil(t, err)
		unknownAccess, err := store.GetByAccess(ctx, info2.GetAccess())
		assert.Nil(t, err, "store.GetAccess")
		assert.Nil(t, unknownAccess, "store.GetAccess")*/
}
