package handlers

import (
	"encoding/json"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func (app *App) GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	username := vars["username"]
	profile := app.DB.GetUserProfile(username)
	if profile.Profile.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonErrorResponse("_", "User not found"))
		return
	}

	userToken := r.Context().Value("user")
	if userToken != nil {
		loggedInUserID := userToken.(*jwt.Token).Claims.(jwt.MapClaims)["UserID"]
		if loggedInUserID != nil {
			profile.Profile.Following = app.DB.IsFollowing(profile.Profile.ID, uint(loggedInUserID.(float64)))
		}
	}

	resp, err := json.Marshal(&profile)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) FollowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	username := vars["username"]
	profile := app.DB.GetUserProfile(username)

	userToken := r.Context().Value("user")
	if userToken == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	loggedInUserID := userToken.(*jwt.Token).Claims.(jwt.MapClaims)["UserID"]
	if loggedInUserID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	app.DB.Follow(profile.Profile.ID, uint(loggedInUserID.(float64)))

	profile.Profile.Following = true
	resp, err := json.Marshal(&profile)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	username := vars["username"]
	profile := app.DB.GetUserProfile(username)

	userToken := r.Context().Value("user")
	if userToken == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	loggedInUserID := userToken.(*jwt.Token).Claims.(jwt.MapClaims)["UserID"]
	if loggedInUserID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	app.DB.Unfollow(profile.Profile.ID, uint(loggedInUserID.(float64)))

	resp, err := json.Marshal(&profile)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}
