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

	user := RegisterUser{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// TODO: Return Http400
		panic("XXX")
	}

	newUser := app.DB.CreateUser(user.User.Username, user.User.Email, user.User.Password)
	newUser.X.NewToken()

	resp, err := json.Marshal(&newUser)
	if err != nil {
		panic(fmt.Errorf("Can not Marshall: %s", err))
	}

	w.Write(resp)
}
