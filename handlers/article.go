package handlers

import (
	"encoding/json"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type ArticleForm struct {
	Article struct {
		Title       string   `json:"title" validate:"required"`
		Description string   `json:"description" validate:"required"`
		Body        string   `json:"body" validate:"required"`
		TagList     []string `json:"tagList"`
	} `json:"article"`
}

func (app *App) ArticleCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	body := ArticleForm{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

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

	article := app.DB.CreateArticle(body.Article.Title, body.Article.Description, body.Article.Body,
		body.Article.TagList, uint(loggedInUserID.(float64)))
	resp, err := json.Marshal(&article)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) ArticleListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	articles := app.DB.ListArticle(r.URL.Query())

	resp, err := json.Marshal(&articles)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) ArticleDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]
	article := app.DB.GetArticleFromSlug(slug)
	if article.Article.Slug == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonErrorNotFoundResponse())
		return
	}

	resp, err := json.Marshal(&article)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)

}

func (app *App) TagsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tags := app.DB.ListTags()
	resp, err := json.Marshal(&tags)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}
