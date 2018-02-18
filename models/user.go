package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"-"`
	Bio      string  `json:"bio"`
	Image    *string `json:"image"`
	Token    string  `gorm:"-" json:"token"`
}
