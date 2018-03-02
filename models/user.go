package models

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
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

type UserResponse struct {
	User User `json:"user"`
}

type MyCustomClaims struct {
	jwt.StandardClaims
	Username string
}

func (db *DB) CreateUser(username, email, password string) *UserResponse {
	password = encryptPassword(password)
	user := User{
		Username: username,
		Email:    email,
		Password: password,
	}
	db.Create(&user)

	return &UserResponse{
		User: user,
	}
}

func (db *DB) UpdateUser(user *User, username, email, password, bio string, image *string) *UserResponse {
	if password != "" {
		password = encryptPassword(password)
	}
	updatedUser := User{
		Username: username,
		Email:    email,
		Password: password,
		Bio:      bio,
		Image:    image,
	}
	db.Model(&user).Updates(&updatedUser)
	return &UserResponse{
		User: updatedUser,
	}
}

func encryptPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Errorf("Encrypt err: %s", err))
	}

	return string(hash)
}

func (db *DB) GetUserFromEmail(email string) *UserResponse {
	user := User{}
	db.Where(&User{Email: email}).First(&user)
	return &UserResponse{
		User: user,
	}
}

func (db *DB) GetUserFromUsername(username string) *UserResponse {
	user := User{}
	db.Where(&User{Username: username}).First(&user)
	return &UserResponse{
		User: user,
	}
}

func GenerateToken(username string) string {
	mySigningKey := []byte(viper.GetString("JWT_SIGNED_KEY"))
	claims := MyCustomClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			Issuer:    "KoYoYo",
		},
		username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		panic(fmt.Errorf("JWT Signed String Error: %s", err))
	}
	return ss
}

func (user *User) NewToken() {
	user.Token = GenerateToken(user.Username)
}

func (user *User) CheckPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return false
	}
	return true
}
