//go:build web

package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func (d *Database) Conn() *gorm.DB {
	return d.db
}

func Open(path string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	d := &Database{db: db}

	if err := db.AutoMigrate(&User{}, &AuditLog{}, &DnsConfig{}, &DnsForwarder{}); err != nil {
		return nil, err
	}

	if err := d.seedDefaultAdmin(); err != nil {
		return nil, err
	}

	if err := d.seedDefaultDnsConfig(); err != nil {
		return nil, err
	}

	if err := d.seedDefaultDnsForwarders(); err != nil {
		return nil, err
	}

	return d, nil
}
