//go:build web

package database

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

func (u *User) SetPassword(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

func (u *User) CheckPassword(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plain)) == nil
}

func (d *Database) seedDefaultAdmin() error {
	var count int64
	d.db.Model(&User{}).Count(&count)
	if count > 0 {
		return nil
	}

	password, err := generateRandomPassword(16)
	if err != nil {
		return err
	}

	admin := &User{Username: "admin"}
	if err := admin.SetPassword(password); err != nil {
		return err
	}

	if err := d.db.Create(admin).Error; err != nil {
		return err
	}

	log.Printf("Default admin user created — password: %s", password)
	return nil
}

func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
