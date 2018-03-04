package handlers

import (
	"github.com/koyoyo/realworld-starter-kit/models"
	"gopkg.in/go-playground/validator.v9"
)

type App struct {
	DB        models.DB
	Validator *validator.Validate
}
