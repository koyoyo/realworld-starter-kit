package models

import (
	"fmt"

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

func encryptPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Errorf("Encrypt err: %s", err))
	}

	return string(hash)
}

func (db *DB) GetUser(email string) *UserResponse {
	user := User{}
	db.Where(&User{Email: email}).First(&user)
	return &UserResponse{
		User: user,
	}
}

func GenerateToken() string {
	mySigningKey := []byte(viper.GetString("JWT_SIGNED_KEY"))
	claims := &jwt.StandardClaims{
		ExpiresAt: 15000,
		Issuer:    "KoYoYo",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		panic(fmt.Errorf("JWT Signed String Error: %s", err))
	}
	return ss
}

func (user *User) NewToken() {
	user.Token = GenerateToken()
}

func (user *User) CheckPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return false
	}
	return true
}
