package models

import (
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
	FavoritesCount uint   `json:"favoritesCount"`
	Author         User   `json:"author"`
	AuthorID       uint
}

type Author struct {
	ID        uint    `json:"-"`
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

type ArticleComment struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	AuthorID  uint
	Author    User
	ArticleID uint
	Article   Article
	Body      string `json:"body"`
}

type Tag struct {
	ID   uint   `json:"-" gorm:"primary_key"`
	Name string `json:"name"`
}

type ArticleResponse struct {
	ID             uint     `json:"-"`
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

type CommentResponse struct {
	ID        uint    `json:"id"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	Body      string  `json:"body"`
	Author    *Author `json:"author"`
}

type CommentResponseJson struct {
	Comment *CommentResponse `json:"comment"`
}

type CommentsResponseJson struct {
	Comments []*CommentResponse `json:"comments"`
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
	return db.PrepareArticleResponse(&article)
}

func (db *DB) UpdateArticle(article *Article, title, description, body string) *ArticleResponseJson {
	if title != "" {
		article.Title = title
		article.Slug = slug.Make(title)
	}

	if description != "" {
		article.Description = description
	}

	if body != "" {
		article.Body = body
	}
	db.Save(&article)

	return db.PrepareArticleResponse(article)
}

func (db *DB) DeleteArticle(article *Article) {
	db.Delete(&article)
}

func (db *DB) listArticle(queries url.Values) (articles []*Article, count uint) {
	sql := db.Preload("Tag").Preload("Author").Order("ID desc")

	if tagQuery, ok := queries["tag"]; ok {
		tag := tagQuery[0]

		sql = sql.Joins("JOIN article_tags ON article_tags.article_id=articles.id").
			Joins("JOIN tags ON article_tags.tag_id=tags.id").
			Where("tags.name = ?", tag)
	}

	if authorQuery, ok := queries["author"]; ok {
		author := authorQuery[0]

		sql = sql.Joins("JOIN users ON users.id=articles.author_id").
			Where("users.username = ?", author)
	}

	if favoritedQuery, ok := queries["favorited"]; ok {
		favorited := favoritedQuery[0]

		var user User
		db.Select("id").Where(&User{Username: favorited}).First(&user)

		if user.ID != 0 {
			sql = sql.Joins("JOIN article_favorites ON article_favorites.article_id=articles.id").
				Where("article_favorites.user_id = ?", user.ID)
		}
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

	sql.Model(&Article{}).Count(&count)
	sql.Offset(offset).Limit(limit).Find(&articles)
	return
}

func (db *DB) ListArticle(queries url.Values) *ArticlesResponseJson {
	articles, count := db.listArticle(queries)
	return db.PrepareArticlesResponse(articles, count)
}

func (db *DB) ListArticleWithUser(queries url.Values, userID uint) *ArticlesResponseJson {
	articles, count := db.listArticle(queries)
	return db.PrepareArticlesResponseWithUser(articles, count, userID)
}

func (db *DB) ListArticleFeed(queries url.Values, userID uint) *ArticlesResponseJson {
	var ids []uint
	db.Find(&Follower{FollowingID: userID}).Pluck("follower_id", &ids)

	sql := db.Where("author_id in (?)", ids).Preload("Tag").Preload("Author").Order("ID desc")

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
	var articles []*Article
	sql.Model(&Article{}).Count(&count)
	sql.Offset(offset).Limit(limit).Find(&articles)
	return db.PrepareArticlesResponseWithUser(articles, count, userID)
}

func (db *DB) CountArticle() uint {
	var count uint
	db.Model(&Article{}).Count(&count)
	return count
}

func (db *DB) GetArticleFromSlug(slug string) *Article {
	var article Article
	db.Preload("Tag").Preload("Author").Where(Article{Slug: slug}).First(&article)
	return &article
}

func (db *DB) GetArticleResponseFromSlug(slug string) *ArticleResponseJson {
	var article Article
	db.Preload("Tag").Preload("Author").Where(Article{Slug: slug}).First(&article)
	return db.PrepareArticleResponse(&article)
}

func (db *DB) PrepareArticleResponse(article *Article) *ArticleResponseJson {
	return &ArticleResponseJson{
		Article: db.PrepareArticle(article),
	}
}

func (db *DB) PrepareArticlesResponse(articles []*Article, count uint) *ArticlesResponseJson {
	var articlesResponse []*ArticleResponse
	for _, article := range articles {
		articlesResponse = append(articlesResponse, db.PrepareArticle(article))
	}

	return &ArticlesResponseJson{
		Articles:      articlesResponse,
		ArticlesCount: count,
	}
}

func (db *DB) PrepareArticlesResponseWithUser(articles []*Article, count uint, userID uint) *ArticlesResponseJson {
	var articlesResponse []*ArticleResponse
	for _, article := range articles {
		article := db.PrepareArticle(article)
		article.Favorited = db.IsFavorite(article.ID, userID)
		article.Author.Following = db.IsFollowing(article.Author.ID, userID)

		articlesResponse = append(articlesResponse, article)
	}

	return &ArticlesResponseJson{
		Articles:      articlesResponse,
		ArticlesCount: count,
	}
}

func (db *DB) PrepareArticle(article *Article) *ArticleResponse {
	tags := []string{}
	for _, tag := range article.Tag {
		tags = append(tags, tag.Name)
	}

	return &ArticleResponse{
		ID:          article.ID,
		CreatedAt:   article.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		UpdatedAt:   article.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Slug:        article.Slug,
		Title:       article.Title,
		Description: article.Description,
		Body:        article.Body,
		Tag:         tags,
		// Favorited: article.Favorited,
		FavoritesCount: article.FavoritesCount,
		Author: &Author{
			ID:        article.Author.ID,
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

func (db *DB) IsFavorite(articleID, userID uint) bool {
	var count uint
	db.Model(&ArticleFavorite{}).Where(&ArticleFavorite{UserID: userID, ArticleID: articleID}).Count(&count)
	return count > 0
}

func (db *DB) FavoriteArticle(articleID, userID uint) (isAlreadyFav bool) {
	var articleFav ArticleFavorite
	if results := db.FirstOrCreate(&articleFav, ArticleFavorite{UserID: userID, ArticleID: articleID}); results.RowsAffected == 0 {
		isAlreadyFav = true
		return
	}

	var countFavorite uint
	db.Model(&ArticleFavorite{}).Where(&ArticleFavorite{ArticleID: articleID}).Count(&countFavorite)
	db.Model(&Article{}).Where(&Article{ID: articleID}).Update("FavoritesCount", countFavorite)
	return
}

func (db *DB) UnfavoriteArticle(articleID, userID uint) (isAlreadyUnfav bool) {
	if results := db.Delete(&ArticleFavorite{}, ArticleFavorite{UserID: userID, ArticleID: articleID}); results.RowsAffected == 0 {
		isAlreadyUnfav = true
		return
	}

	var countFavorite uint
	db.Model(&ArticleFavorite{}).Where(&ArticleFavorite{ArticleID: articleID}).Count(&countFavorite)
	db.Model(&Article{}).Where(&Article{ID: articleID}).Update("FavoritesCount", countFavorite)
	return
}

func (db *DB) AddArticleComment(article *Article, userID uint, body string) *CommentResponseJson {
	comment := &ArticleComment{
		AuthorID:  userID,
		ArticleID: article.ID,
		Body:      body,
	}
	db.Create(&comment)

	var author User
	db.First(&author, userID)
	comment.Author = author
	return db.PrepareCommentResponse(comment)
}

func (db *DB) listArticleComment(articleID uint) (comments []*ArticleComment) {
	db.Where(&ArticleComment{ArticleID: articleID}).Order("ID desc").Find(&comments)
	return
}

func (db *DB) ListArticleComment(articleID uint) *CommentsResponseJson {
	comments := db.listArticleComment(articleID)
	return db.PrepareCommentsResponse(comments)
}

func (db *DB) ListArticleCommentWithUser(articleID uint, userID uint) *CommentsResponseJson {
	comments := db.listArticleComment(articleID)
	return db.PrepareCommentsResponseWithUser(comments, userID)
}

func (db *DB) GetArticleComment(commentID uint, articleSlug string) *ArticleComment {
	var comment ArticleComment
	db.Debug().Model(&ArticleComment{}).Where(&ArticleComment{ID: commentID}).
		Joins("JOIN articles ON articles.ID=article_comments.article_id").
		Where("articles.slug = ?", articleSlug).
		First(&comment)
	return &comment
}

func (db *DB) DeleteArticleComment(comment *ArticleComment) {
	db.Delete(&comment)
}

func (db *DB) PrepareCommentResponse(comment *ArticleComment) *CommentResponseJson {
	return &CommentResponseJson{
		Comment: db.PrepareComment(comment),
	}
}

func (db *DB) PrepareCommentsResponse(comments []*ArticleComment) *CommentsResponseJson {
	var commentsResponse []*CommentResponse
	for _, comment := range comments {
		commentsResponse = append(commentsResponse, db.PrepareComment(comment))
	}

	return &CommentsResponseJson{
		Comments: commentsResponse,
	}
}

func (db *DB) PrepareCommentsResponseWithUser(comments []*ArticleComment, userID uint) *CommentsResponseJson {
	var commentsResponse []*CommentResponse
	for _, comment := range comments {
		comment := db.PrepareComment(comment)
		comment.Author.Following = db.IsFollowing(comment.Author.ID, userID)

		commentsResponse = append(commentsResponse, comment)
	}

	return &CommentsResponseJson{
		Comments: commentsResponse,
	}
}

func (db *DB) PrepareComment(comment *ArticleComment) *CommentResponse {
	return &CommentResponse{
		ID:        comment.ID,
		CreatedAt: comment.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		UpdatedAt: comment.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Body:      comment.Body,
		Author: &Author{
			ID:        comment.Author.ID,
			Username:  comment.Author.Username,
			Bio:       comment.Author.Bio,
			Image:     comment.Author.Image,
			Following: false,
		},
	}
}
