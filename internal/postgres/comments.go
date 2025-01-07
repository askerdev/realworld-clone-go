package postgres

import (
	"context"
	"strings"

	"github.com/askerdev/realworld-clone-go/internal/domain/entity"
)

type SelectCommentsParams struct {
	CommentID   *uint64
	ArticleSlug string
	UserID      *uint64
}

func (s *Storage) SelectComments(
	ctx context.Context,
	params *SelectCommentsParams,
) ([]*entity.Comment, error) {
	conditionalSelect := ""
	conditionalJoin := ""
	conditionalWhere := []string{}
	where := ""
	args := NewArgs()

	if params.CommentID != nil {
		args.Append(*params.CommentID)
		conditionalWhere = append(conditionalWhere, "c.id = "+args.Placeholder)
	}

	if params.UserID != nil {
		args.Append(*params.UserID)
		conditionalSelect = ` , s.user_id AS subscriber_id `
		conditionalJoin = ` LEFT JOIN subscriptions s ON s.profile_id = c.author_id AND s.user_id = ` + args.Placeholder
	}

	args.Append(params.ArticleSlug)
	conditionalWhere = append(conditionalWhere, `c.article_id IN (
      SELECT id AS aid FROM articles WHERE slug = `+args.Placeholder+`)`)

	if len(conditionalWhere) > 0 {
		where = " WHERE "
	}

	query := `
    SELECT 
      c.id, c.body, c.author_id, c.article_id, c.created_at, c.updated_at,
      u.username AS user_username, u.bio AS user_bio, u.image AS user_image
    ` +
		conditionalSelect +
		`FROM
      comments c
		  INNER JOIN users u ON u.id = c.author_id
		` + conditionalJoin + where + strings.Join(conditionalWhere, " AND ")

	rows, err := s.db.QueryxContext(ctx, query, args.Values...)
	if err != nil {
		return nil, err
	}

	comments := []*entity.Comment{}

	for rows.Next() {
		row := &CommentRow{}
		if err := rows.StructScan(row); err != nil {
			return nil, err
		}
		comment := convertCommentRowToComment(row)
		if row.SubscriberID != nil && params.UserID != nil {
			comment.Author.Following = *row.SubscriberID == *params.UserID
		}
		comments = append(comments, comment)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return comments, nil
}

type InsertCommentParams struct {
	ArticleSlug string
	UserID      uint64
	Body        string
}

func (s *Storage) InsertComment(
	ctx context.Context,
	params *InsertCommentParams,
) (*entity.Comment, error) {
	const query = `
    INSERT INTO comments
      (body, author_id, article_id)
    VALUES
      ($1, $2, (SELECT id FROM articles WHERE slug = $3 LIMIT 1))
    RETURNING comments.id`

	row := s.db.QueryRowxContext(ctx, query, params.Body, params.UserID, params.ArticleSlug)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var id uint64
	if err := row.Scan(&id); err != nil {
		return nil, err
	}

	comments, err := s.SelectComments(ctx, &SelectCommentsParams{
		CommentID:   &id,
		UserID:      &params.UserID,
		ArticleSlug: params.ArticleSlug,
	})
	if err != nil {
		return nil, err
	}

	if len(comments) < 1 {
		panic("inserted comment not found")
	}

	return comments[0], nil
}

type DeleteCommentParams struct {
	CommentID   uint64
	ArticleSlug string
	UserID      uint64
}

func (s *Storage) DeleteComment(
	ctx context.Context,
	params *DeleteCommentParams,
) error {
	const query = `
    DELETE FROM comments
    WHERE id = $1 AND author_id = $2 AND article_id IN (
      SELECT id FROM articles WHERE slug = $3 LIMIT 1
    )`

	_, err := s.db.ExecContext(ctx, query, params.CommentID, params.UserID, params.ArticleSlug)
	return err
}
