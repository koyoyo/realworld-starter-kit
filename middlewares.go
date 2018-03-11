package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

func customFromAuthHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil // No error, just no token
	}

	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "token" {
		return "", errors.New("Authorization header format must be Token {token}")
	}

	return authHeaderParts[1], nil
}

var JwtRequiredMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("JWT_SIGNED_KEY")), nil
	},
	SigningMethod: jwt.SigningMethodHS256,
	Extractor:     customFromAuthHeader,
})

var JwtOptionalMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("JWT_SIGNED_KEY")), nil
	},
	SigningMethod:       jwt.SigningMethodHS256,
	Extractor:           customFromAuthHeader,
	CredentialsOptional: true,
})
