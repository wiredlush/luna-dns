//go:build web

package web

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Setting struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

func OpenDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Setting{}); err != nil {
		return nil, err
	}

	return db, nil
}
