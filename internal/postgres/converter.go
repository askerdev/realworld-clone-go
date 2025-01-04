package postgres

import "github.com/askerdev/realworld-clone-go/internal/domain/entity"

func convertArticleRowWithTagAndUserToDomainArticle(
	articleRow *ArticleRowWithTagAndUser,
) *entity.Article {
	article := &entity.Article{
		ID:             articleRow.ID,
		Slug:           articleRow.Slug,
		Title:          articleRow.Title,
		Description:    articleRow.Description,
		Body:           articleRow.Body,
		TagList:        []string{},
		FavoritesCount: articleRow.FavoritesCount,
		CreatedAt:      articleRow.CreatedAt,
		UpdatedAt:      articleRow.UpdatedAt,
		Author: &entity.Profile{
			ID:       articleRow.UserID,
			Username: articleRow.UserUsername,
			Bio:      articleRow.UserBio,
			Image:    articleRow.UserImage,
		},
	}
	if articleRow.Tag.Valid {
		article.TagList = append(article.TagList, articleRow.Tag.String)
	}

	return article
}
