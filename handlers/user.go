package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RegisterUser struct {
	User struct {
		Username string `json: username`
		Email    string `json: email`
		Password string `json: password`
	} `json: user`
}

func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	// TODO: Validate Form

	body := RegisterUser{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		// TODO: Return Http400
		panic("XXX")
	}

	newUser := app.DB.CreateUser(body.User.Username, body.User.Email, body.User.Password)
	newUser.User.NewToken()

	resp, err := json.Marshal(&newUser)
	if err != nil {
		panic(fmt.Errorf("Can not Marshall: %s", err))
	}

	w.Write(resp)
}
