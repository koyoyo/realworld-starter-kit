package models

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gosimple/slug"
)

type Article struct {
	ID        uint       `gorm:"primary_key"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`

	Slug           string `json:"slug"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Body           string `json:"body"`
	Tag            []Tag  `gorm:"many2many:article_tags;" json:"tagList"`
	Favorited      bool   `gorm:"-" json:"favorited"`
	FavoritesCount uint   `gorm:"-" json:"favoritesCount"`
	Author         User   `json:"author"`
	AuthorID       uint
}

type Author struct {
	Username  string  `json:"username"`
	Bio       string  `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

type ArticleFavorite struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time

	UserID    uint `gorm:"unique_index:favorite"`
	User      User
	ArticleID uint `gorm:"unique_index:favorite"`
	Article   Article
}

type Tag struct {
	ID   uint   `json:"-" gorm:"primary_key"`
	Name string `json:"name"`
}

type ArticleResponse struct {
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Body           string   `json:"body"`
	Tag            []string `json:"tagList"`
	Favorited      bool     `json:"favorited"`
	FavoritesCount uint     `json:"favoritesCount"`
	Author         *Author  `json:"author"`
}

type ArticleResponseJson struct {
	Article *ArticleResponse `json:"article"`
}

type ArticlesResponseJson struct {
	Articles      []*ArticleResponse `json:"articles"`
	ArticlesCount uint               `json:"articlesCount"`
}

type TagResponse struct {
	Tags []string `json:"tags"`
}

func (db *DB) CreateArticle(title, description, body string, tagList []string, userID uint) *ArticleResponseJson {
	article := Article{
		Title:       title,
		Slug:        slug.Make(title),
		Description: description,
		Body:        body,
		AuthorID:    userID,
	}

	var tags []Tag
	for _, tagName := range tagList {
		var tag Tag
		db.Where(Tag{Name: tagName}).FirstOrCreate(&tag)
		tags = append(tags, tag)
	}
	article.Tag = tags

	db.Create(&article)

	var author User
	db.First(&author, userID)
	article.Author = author
	return db.PrepareArticleResponse(article)
}

func (db *DB) ListArticle(queries url.Values) *ArticlesResponseJson {
	var articles []Article

	sql := db.Preload("Tag").Preload("Author").Order("ID desc")

	var tag string
	if tagQuery, ok := queries["tag"]; ok {
		tag = tagQuery[0]

		sql = sql.Joins("JOIN article_tags ON article_tags.article_id=articles.id").
			Joins("JOIN tags ON article_tags.tag_id=tags.id").
			Where("tags.name = ?", tag)
	}

	var author string
	if authorQuery, ok := queries["author"]; ok {
		author = authorQuery[0]

		sql = sql.Joins("JOIN users ON users.id=articles.author_id").
			Where("users.username = ?", author)
	}
	if favorited, ok := queries["favorited"]; ok {
		fmt.Println("Has favorited", favorited)
	}

	limit := 20
	if limitStr, ok := queries["limit"]; ok {
		if limitTmp, err := strconv.Atoi(limitStr[0]); err == nil {
			limit = limitTmp
		}
	}
	var offset int
	if offsetStr, ok := queries["offset"]; ok {
		if offsetTmp, err := strconv.Atoi(offsetStr[0]); err == nil {
			offset = offsetTmp
		}
	}

	var count uint
	sql.Model(&Article{}).Count(&count)
	sql.Offset(offset).Limit(limit).Find(&articles)

	return db.PrepareArticlesResponse(articles, count)
}

func (db *DB) CountArticle() uint {
	var count uint
	db.Model(&Article{}).Count(&count)
	return count
}

func (db *DB) GetArticleFromSlug(slug string) *ArticleResponseJson {
	var article Article
	db.Preload("Tag").Where(Article{Slug: slug}).First(&article)
	return db.PrepareArticleResponse(article)
}

func (db *DB) PrepareArticleResponse(article Article) *ArticleResponseJson {
	return &ArticleResponseJson{
		Article: db.PrepareArticle(article),
	}
}

func (db *DB) PrepareArticlesResponse(articles []Article, count uint) *ArticlesResponseJson {
	var articlesResponse []*ArticleResponse
	for _, article := range articles {
		articlesResponse = append(articlesResponse, db.PrepareArticle(article))
	}

	return &ArticlesResponseJson{
		Articles:      articlesResponse,
		ArticlesCount: count,
	}
}

func (db *DB) PrepareArticle(article Article) *ArticleResponse {
	tags := []string{}
	for _, tag := range article.Tag {
		tags = append(tags, tag.Name)
	}

	return &ArticleResponse{
		CreatedAt:   article.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		UpdatedAt:   article.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		Slug:        article.Slug,
		Title:       article.Title,
		Description: article.Description,
		Body:        article.Body,
		Tag:         tags,
		// Favorited: article.Favorited,
		// FavoritesCount: article.FavoritesCount,
		Author: &Author{
			Username:  article.Author.Username,
			Bio:       article.Author.Bio,
			Image:     article.Author.Image,
			Following: false,
		},
	}
}

func (db *DB) ListTags() *TagResponse {
	var tags []string
	db.Model(&Tag{}).Pluck("name", &tags)
	return &TagResponse{
		Tags: tags,
	}
}
