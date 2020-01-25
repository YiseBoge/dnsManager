package models

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"time"
)

type DomainName struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name     string
	Address  string
	LastRead time.Time
}
