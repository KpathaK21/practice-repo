package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	Username                string `gorm:"unique;not null"`
	Email                   string `gorm:"unique;not null"`
	Password                string `gorm:"not null"`
	VerificationCode        string
	VerificationCodeCreated int64 // Unix timestamp when the code was created
	IsVerified              bool
}

func (u *User) SetPassword(pw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(pw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pw))
	return err == nil
}
