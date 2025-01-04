package postgres

import (
	"context"
	"strconv"
	"strings"

	"github.com/askerdev/realworld-clone-go/internal/domain/entity"
	"github.com/guregu/null/v5"
)

type CreateArticleParams struct {
	AuthorID    uint64
	Slug        string
	Title       string
	Description string
	Body        string
	TagList     []string
}

func (s *Storage) CreateArticle(
	ctx context.Context,
	params *CreateArticleParams,
) (*entity.Article, error) {
	const insertArticleQuery = `
    INSERT INTO articles
      (slug, title, description, body, author_id)
    VALUES
      ($1, $2, $3, $4, $5)
    RETURNING *`

	row := s.db.QueryRowxContext(
		ctx,
		insertArticleQuery,
		params.Slug, params.Title, params.Description,
		params.Body, params.AuthorID,
	)
	articleRow := &ArticleRow{}
	if err := row.StructScan(articleRow); err != nil {
		return nil, err
	}

	profile, err := s.SelectProfileByID(ctx, params.AuthorID, nil)
	if err != nil {
		return nil, err
	}

	err = s.InsertTags(ctx, articleRow.ID, params.TagList)
	if err != nil {
		return nil, err
	}

	article := entity.Article{
		ID:             articleRow.ID,
		Slug:           articleRow.Slug,
		Title:          articleRow.Title,
		Body:           articleRow.Body,
		Description:    articleRow.Description,
		FavoritesCount: articleRow.FavoritesCount,
		CreatedAt:      articleRow.CreatedAt,
		UpdatedAt:      articleRow.UpdatedAt,
		TagList:        params.TagList,
		Author:         profile,
		Favorited:      false,
	}

	return &article, nil
}

type SelectArticlesParams struct {
	UserID              *uint64
	Tag                 null.String
	AuthorUsername      null.String
	FavoritedByUsername null.String
	Limit               null.Int
	Offset              null.Int
}

func (s *Storage) SelectArticles(
	ctx context.Context,
	params *SelectArticlesParams,
) ([]*entity.Article, error) {
	subscriberJoin := ""
	end := ""
	where := []string{}
	args := NewArgs()

	if params.UserID != nil {
		args.Append(*params.UserID)
		subscriberJoin += " LEFT JOIN subscriptions s ON s.profile_id = author_id AND s.user_id = " + args.Placeholder + " "
	}

	if params.Tag.Valid {
		args.Append(params.Tag.String)
		where = append(
			where,
			`id IN (SELECT article_id FROM tags WHERE value = `+args.Placeholder+")",
		)
	}

	if params.AuthorUsername.Valid {
		args.Append(params.AuthorUsername.String)
		where = append(where, "u.username = "+args.Placeholder)
	}

	if params.Limit.Valid && params.Limit.Int64 > 0 && params.Limit.Int64 <= 20 {
		end += " LIMIT " + strconv.FormatUint(uint64(params.Limit.Int64), 10)
	} else {
		end += " LIMIT 20"
	}

	if params.Offset.Valid && params.Offset.Int64 > 0 {
		end += " OFFSET " + strconv.FormatUint(uint64(params.Offset.Int64), 10)
	}

	whereStart := " "
	if len(where) > 0 {
		whereStart = " WHERE "
	}

	selectFollowingSubscriber := " "
	if len(subscriberJoin) > 0 {
		selectFollowingSubscriber = ", s.user_id AS subscriber_id "
	}

	query := `
    SELECT
      t.value AS article_tag,
      id, slug, title, description, body, favorites_count, created_at, updated_at, author_id,
      u_user_id AS user_id, user_bio, user_username, user_image ` + selectFollowingSubscriber +
		`FROM (
      SELECT 
        a.*,
        u.id AS u_user_id, u.bio AS user_bio, u.username AS user_username, u.image AS user_image
      FROM articles a
      INNER JOIN users u ON u.id = a.author_id` + whereStart + strings.Join(where, " AND ") + `
      ORDER BY a.created_at DESC` + end + `
    )
    LEFT JOIN tags t ON t.article_id = id
    ` + subscriberJoin + ` ORDER BY created_at DESC`

	rows, err := s.db.QueryxContext(ctx, query, args.Values...)
	if err != nil {
		return nil, err
	}

	articles := []*entity.Article{}
	var article *entity.Article
	for rows.Next() {
		articleRow := &ArticleRowWithTagAndUser{}
		if err := rows.StructScan(articleRow); err != nil {
			rows.Close()
			return nil, err
		}
		if article == nil {
			article = convertArticleRowWithTagAndUserToDomainArticle(articleRow)
			if articleRow.SubscriberID != nil && params.UserID != nil {
				article.Author.Following = *params.UserID == *articleRow.SubscriberID
			}
		} else if articleRow.ID == article.ID && articleRow.Tag.Valid {
			article.TagList = append(article.TagList, articleRow.Tag.String)
		} else {
			articles = append(articles, article)
			article = convertArticleRowWithTagAndUserToDomainArticle(articleRow)
			if articleRow.SubscriberID != nil && params.UserID != nil {
				article.Author.Following = *params.UserID == *articleRow.SubscriberID
			}
		}
	}
	if article != nil {
		articles = append(articles, article)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return articles, nil
}
