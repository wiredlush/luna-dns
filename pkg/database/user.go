//go:build web

package database

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func (d *Database) FindByUsername(username string) (*User, error) {
	var user User
	if err := d.db.Session(&gorm.Session{Logger: logger.Discard}).
		Where("username = ?", username).
		First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *Database) ListUsers() ([]User, error) {
	var users []User
	if err := d.db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (d *Database) CreateUser(username, password string) (*User, error) {
	user := &User{Username: username}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	if err := d.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (d *Database) FindByID(id uint) (*User, error) {
	var user User
	if err := d.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *Database) DeleteUser(id uint) error {
	return d.db.Unscoped().Delete(&User{}, id).Error
}

func (d *Database) CountUsers() (int64, error) {
	var count int64
	err := d.db.Model(&User{}).Count(&count).Error
	return count, err
}

func (d *Database) UpdatePassword(username, password string) error {
	user, err := d.FindByUsername(username)
	if err != nil {
		return err
	}

	if err := user.SetPassword(password); err != nil {
		return err
	}

	return d.db.Save(user).Error
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
