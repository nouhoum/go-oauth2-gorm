package gooauth2gorm

import (
	"time"

	"gorm.io/gorm"
)

type Client struct {
	ID        string
	Secret    string
	Domain    string
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ClientStore struct {
}

func New() *ClientStore {
	return nil
}
