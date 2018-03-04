package handlers

import (
	"encoding/json"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

type RegisterUser struct {
	User struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	} `json:"user"`
}

type LoginUser struct {
	User struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	} `json:"user"`
}

type UpdateUser struct {
	User struct {
		Username string  `json:"username"`
		Email    string  `json:"email" validate:"email"`
		Password string  `json:"password"`
		Bio      string  `json:"bio"`
		Image    *string `json:"image" validate:"url"`
	} `json:"user"`
}

func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	body := RegisterUser{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	err = app.Validator.Struct(body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponseFromValidator(err))
		return
	}

	newUser := app.DB.CreateUser(body.User.Username, body.User.Email, body.User.Password)
	newUser.User.NewToken()

	resp, err := json.Marshal(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	body := LoginUser{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	err = app.Validator.Struct(body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponseFromValidator(err))
		return
	}

	user := app.DB.GetUserFromEmail(body.User.Email)
	if user.User.ID == 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("email", "is invalid"))
		return
	}

	isMatch := user.User.CheckPassword(body.User.Password)
	if !isMatch {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("password", "is invalid"))
		return
	}

	user.User.NewToken()
	resp, err := json.Marshal(&user)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userToken := r.Context().Value("user")
	username := userToken.(*jwt.Token).Claims.(jwt.MapClaims)["Username"]
	if username == nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", "Can't extract user token."))
		return
	}

	user := app.DB.GetUserFromUsername(username.(string))
	user.User.Token = userToken.(*jwt.Token).Raw
	resp, err := json.Marshal(&user)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body := UpdateUser{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	err = app.Validator.Struct(body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponseFromValidator(err))
		return
	}

	userToken := r.Context().Value("user")
	username := userToken.(*jwt.Token).Claims.(jwt.MapClaims)["Username"]
	if username == nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", "Can't extract user token."))
		return
	}

	user := app.DB.GetUserFromUsername(username.(string))
	updatedUser := app.DB.UpdateUser(&user.User, body.User.Username, body.User.Email, body.User.Password, body.User.Bio,
		body.User.Image)
	updatedUser.User.Token = userToken.(*jwt.Token).Raw

	resp, err := json.Marshal(&updatedUser)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}
