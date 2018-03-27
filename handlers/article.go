package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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

type CommentForm struct {
	Comment struct {
		Body string `json:"body" validate:"required"`
	} `json:"comment"`
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

func (app *App) ArticleFeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	articles := app.DB.ListArticleFeed(r.URL.Query(), uint(loggedInUserID.(float64)))

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
	article := app.DB.GetArticleResponseFromSlug(slug)
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

func (app *App) ArticleUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	body := ArticleForm{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	article := app.DB.GetArticleFromSlug(slug)
	if article.Slug == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonErrorNotFoundResponse())
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

	if article.AuthorID != uint(loggedInUserID.(float64)) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	articleResponse := app.DB.UpdateArticle(article, body.Article.Title, body.Article.Description, body.Article.Body)
	resp, err := json.Marshal(&articleResponse)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) ArticleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]
	article := app.DB.GetArticleFromSlug(slug)
	if article.Slug == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonErrorNotFoundResponse())
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

	if article.AuthorID != uint(loggedInUserID.(float64)) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	app.DB.DeleteArticle(article)
	w.WriteHeader(http.StatusNoContent)
	return
}

func (app *App) ArticleFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]
	article := app.DB.GetArticleResponseFromSlug(slug)

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
	article := app.DB.GetArticleResponseFromSlug(slug)

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

func (app *App) ArticleCommentAddHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	body := CommentForm{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	article := app.DB.GetArticleFromSlug(slug)
	if article.Slug == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonErrorNotFoundResponse())
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

	comment := app.DB.AddArticleComment(article, uint(loggedInUserID.(float64)), body.Comment.Body)
	resp, err := json.Marshal(&comment)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) ArticleCommentListHandler(w http.ResponseWriter, r *http.Request) {
	var comments *models.CommentsResponseJson

	vars := mux.Vars(r)
	slug := vars["slug"]
	article := app.DB.GetArticleFromSlug(slug)
	if article.Slug == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonErrorNotFoundResponse())
		return
	}

	if userToken := r.Context().Value("user"); userToken != nil {
		if loggedInUserID := userToken.(*jwt.Token).Claims.(jwt.MapClaims)["UserID"]; loggedInUserID != nil {
			comments = app.DB.ListArticleCommentWithUser(article.ID, uint(loggedInUserID.(float64)))
		} else {
			comments = app.DB.ListArticleComment(article.ID)
		}
	} else {
		comments = app.DB.ListArticleComment(article.ID)
	}

	resp, err := json.Marshal(&comments)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	w.Write(resp)
}

func (app *App) ArticleCommentDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	slug := vars["slug"]
	commentID, err := strconv.Atoi(vars["commentID"])
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(JsonErrorResponse("_", err.Error()))
		return
	}

	comment := app.DB.GetArticleComment(uint(commentID), slug)
	if comment.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write(JsonErrorNotFoundResponse())
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

	if comment.AuthorID != uint(loggedInUserID.(float64)) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	app.DB.DeleteArticleComment(comment)
	w.WriteHeader(http.StatusNoContent)
	return
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
