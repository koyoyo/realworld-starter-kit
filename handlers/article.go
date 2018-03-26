package handlers

import (
	"encoding/json"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/koyoyo/realworld-starter-kit/models"
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

	var articles *models.ArticlesResponseJson

	if userToken := r.Context().Value("user"); userToken != nil {
		if loggedInUserID := userToken.(*jwt.Token).Claims.(jwt.MapClaims)["UserID"]; loggedInUserID != nil {
			articles = app.DB.ListArticleWithUser(r.URL.Query(), uint(loggedInUserID.(float64)))
		} else {
			articles = app.DB.ListArticle(r.URL.Query())
		}
	} else {
		articles = app.DB.ListArticle(r.URL.Query())
	}

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

	userToken := r.Context().Value("user")
	if userToken != nil {
		loggedInUserID := userToken.(*jwt.Token).Claims.(jwt.MapClaims)["UserID"]
		if loggedInUserID != nil {
			article.Article.Favorited = app.DB.IsFavorite(article.Article.ID, uint(loggedInUserID.(float64)))
			article.Article.Author.Following = app.DB.IsFollowing(article.Article.Author.ID, uint(loggedInUserID.(float64)))
		}
	}

	resp, err := json.Marshal(&article)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)

}

func (app *App) ArticleFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]
	article := app.DB.GetArticleFromSlug(slug)

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

	isAlreadyFav := app.DB.FavoriteArticle(article.Article.ID, uint(loggedInUserID.(float64)))
	if !isAlreadyFav {
		article.Article.FavoritesCount++
	}

	article.Article.Favorited = true
	resp, err := json.Marshal(&article)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) ArticleUnfavoriteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]
	article := app.DB.GetArticleFromSlug(slug)

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

	isAlreadyUnfav := app.DB.UnfavoriteArticle(article.Article.ID, uint(loggedInUserID.(float64)))
	if !isAlreadyUnfav {
		article.Article.FavoritesCount--
	}

	article.Article.Favorited = false
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