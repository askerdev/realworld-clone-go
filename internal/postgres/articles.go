package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

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
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	const insertArticleQuery = `
    INSERT INTO articles
      (slug, title, description, body, author_id)
    VALUES
      ($1, $2, $3, $4, $5)
    RETURNING *`

	row := tx.QueryRowxContext(
		ctx,
		insertArticleQuery,
		params.Slug, params.Title, params.Description,
		params.Body, params.AuthorID,
	)
	articleRow := &ArticleRow{}
	if err := row.StructScan(articleRow); err != nil {
		tx.Rollback()
		return nil, err
	}

	profile, err := s.selectProfileByID(ctx, tx, params.AuthorID, nil)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = s.saveTags(ctx, tx, articleRow.ID, params.TagList)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	article := &entity.Article{
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

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return article, nil
}

type UpdateArticleParams struct {
	OriginalSlug string
	Slug         null.String
	Title        null.String
	Description  null.String
	Body         null.String
}

func (s *Storage) UpdateArticle(
	ctx context.Context,
	params *UpdateArticleParams,
) error {
	fields := []string{}
	args := NewArgs()

	if params.Title.Valid && params.Slug.Valid {
		args.Append(params.Slug.String)
		fields = append(fields, "slug = "+args.Placeholder)
		args.Append(params.Title.String)
		fields = append(fields, "title = "+args.Placeholder)
	}

	if params.Description.Valid {
		args.Append(params.Description.String)
		fields = append(fields, "description = "+args.Placeholder)
	}

	if params.Body.Valid {
		args.Append(params.Body.String)
		fields = append(fields, "body = "+args.Placeholder)
	}

	if len(fields) == 0 {
		return nil
	}

	args.Append(time.Now())
	fields = append(fields, "updated_at = "+args.Placeholder)

	args.Append(params.OriginalSlug)

	updateArticleQuery := `
    UPDATE articles SET ` + strings.Join(fields, ", ") +
		` WHERE slug = ` + args.Placeholder

	res, err := s.db.ExecContext(
		ctx,
		updateArticleQuery,
		args.Values...,
	)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) RemoveArticle(
	ctx context.Context,
	slug string,
	authorID uint64,
) error {
	const query = `DELETE FROM articles CASCADE WHERE slug = $1 AND author_id = $2`

	res, err := s.db.ExecContext(
		ctx,
		query,
		slug, authorID,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type SelectArticlesParams struct {
	UserID              *uint64
	Feed                bool
	Tag                 null.String
	AuthorUsername      null.String
	FavoritedByUsername null.String
	Slug                null.String
	Limit               null.Int
	Offset              null.Int
}

func (s *Storage) SelectArticles(
	ctx context.Context,
	params *SelectArticlesParams,
) ([]*entity.Article, uint, error) {
	conditionalJoin := ""
	authenticatedJoin := ""
	end := ""
	where := []string{}
	args := NewArgs()

	if params.FavoritedByUsername.Valid {
		args.Append(params.FavoritedByUsername.String)
		conditionalJoin += ` INNER JOIN favorites_articles_rel farbyu ON farbyu.article_id = a.id AND farbyu.user_id IN (
      SELECT id FROM users WHERE username = ` + args.Placeholder + `)`
	}

	if params.UserID != nil {
		args.Append(*params.UserID)
		joinType := "LEFT"
		if params.Feed {
			joinType = "INNER"
		}
		authenticatedJoin += " " + joinType + " JOIN subscriptions s ON s.profile_id = author_id AND s.user_id = " + args.Placeholder + `
      LEFT JOIN favorites_articles_rel far ON far.article_id = a.id AND far.user_id = ` + args.Placeholder
	}

	if params.Slug.Valid {
		args.Append(params.Slug.String)
		where = append(where, `a.slug = `+args.Placeholder)
	}

	if params.Tag.Valid {
		args.Append(params.Tag.String)
		where = append(
			where,
			`a.id IN (
        SELECT tar.article_id as article_id FROM tags
        JOIN tags_articles_rel tar ON tar.tag_id = tags.id
        WHERE value = `+args.Placeholder+")",
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

	authenticatedSelect := " "
	if len(authenticatedJoin) > 0 {
		authenticatedSelect += ", far.user_id AS favorited_by_id, s.user_id AS subscriber_id "
	}

	query := `
    SELECT
      t.value AS article_tag,
      a.id, a.slug, a.title, a.description, a.body, a.favorites_count, a.created_at, a.updated_at, a.author_id,
      u_user_id AS user_id, user_bio, user_username, user_image,
      articles_count
    ` + authenticatedSelect +
		`FROM (
      SELECT 
        a.*,
        u.id AS u_user_id, u.bio AS user_bio, u.username AS user_username, u.image AS user_image,
        COUNT(a.*) AS articles_count
      FROM articles a
      INNER JOIN users u ON u.id = a.author_id ` + whereStart + strings.Join(where, " AND ") + `
      GROUP BY a.id, u.id
      ORDER BY a.created_at DESC` + end + `
    ) a
    LEFT JOIN tags_articles_rel tar ON tar.article_id = id
    LEFT JOIN tags t ON tar.tag_id = t.id
    ` + conditionalJoin + authenticatedJoin + ` ORDER BY created_at DESC, t.value`

	rows, err := s.db.QueryxContext(ctx, query, args.Values...)
	if err != nil {
		return nil, 0, err
	}

	var articlesCount uint
	articles := []*entity.Article{}
	var article *entity.Article
	for rows.Next() {
		articleRow := &ArticleRowWithTagAndUser{}
		if err := rows.StructScan(articleRow); err != nil {
			rows.Close()
			return nil, 0, err
		}
		articlesCount = uint(articleRow.ArticlesCount)
		if article == nil {
			article = convertArticleRowWithTagAndUserToDomainArticle(articleRow)
			if params.UserID != nil {
				if articleRow.SubscriberID != nil {
					article.Author.Following = *params.UserID == *articleRow.SubscriberID
				}
				if articleRow.FavoritedByID != nil {
					article.Favorited = *params.UserID == *articleRow.FavoritedByID
				}
			}
		} else if articleRow.ID == article.ID && articleRow.Tag.Valid {
			article.TagList = append(article.TagList, articleRow.Tag.String)
		} else {
			articles = append(articles, article)
			article = convertArticleRowWithTagAndUserToDomainArticle(articleRow)
			if params.UserID != nil {
				if articleRow.SubscriberID != nil {
					article.Author.Following = *params.UserID == *articleRow.SubscriberID
				}
				if articleRow.FavoritedByID != nil {
					article.Favorited = *params.UserID == *articleRow.FavoritedByID
				}
			}
		}
	}
	if article != nil {
		articles = append(articles, article)
	}

	if err := rows.Close(); err != nil {
		return nil, 0, err
	}

	return articles, articlesCount, nil
}
