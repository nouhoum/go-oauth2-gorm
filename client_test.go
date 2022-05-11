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

	//Remove by code
	clientID := client.GetID()
	err = store.RemoveByID(ctx, clientID)
	assert.Nil(t, err)
	unknownClient, err := store.GetByID(ctx, clientID)
	assert.Nil(t, err, "store.GetByID")
	assert.Nil(t, unknownClient, "store.GetByID")
}
