package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/jinzhu/gorm"
)

type Role int

const (
	Admin Role = iota
	Employee
)

func (r Role) String() string {
	switch r {
	case Admin:
		return "admin"
	case Employee:
		return "employee"
	default:
		return "unknown"
	}
}

type User struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" sql:"index"`

	Username string `json:"username" gorm:"unique_index;not null"`
	Email    string `json:"email" gorm:"unique_index;not null"`
	Password string `json:"password" gorm:"not null"`
	Role     Role   `json:"role" gorm:"default:1"` // 0: admin, 1: employee

	// 关联内容
	Contents []Content `json:"contents" gorm:"foreignkey:UserID"`
}

// HashPassword hashes the password using bcrypt
func (u *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

// CheckPassword checks if the provided password matches the hashed password
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// TableName sets the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to hash password before creating user
func (u *User) BeforeCreate(scope *gorm.Scope) error {
	if u.Password != "" {
		return u.HashPassword(u.Password)
	}
	return nil
}

// BeforeUpdate hook to hash password before updating user
func (u *User) BeforeUpdate(scope *gorm.Scope) error {
	if u.Password != "" {
		return u.HashPassword(u.Password)
	}
	return nil
}