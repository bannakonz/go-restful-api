package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Name     string `gorm:"not null"`
	Avatar   string
	Role     string `gorm:"default:Member;not null"`
}

func (u *User) Promote() {
	u.Role = "Editor"
}

func (u *User) Demote() {
	u.Role = "Member"
}

//func (u *User) BeforeSave(tx *gorm.DB) error {
//	if u.Password == "" {
//		return nil
//	}
//
//	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
//	if err != nil {
//		return err
//	}
//
//	u.Password = string(hash)
//	return nil
//}

func (u *User) GenerateEncryptedPassword() (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(u.Password),
		14,
	)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}
