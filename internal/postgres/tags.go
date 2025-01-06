package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
)

func (s *Storage) saveTags(
	ctx context.Context,
	tx *sqlx.Tx,
	articleID uint64,
	tags []string,
) error {
	if len(tags) == 0 {
		return nil
	}

	insertTagsFields := []string{}
	insertTagsArgs := NewArgs()

	for _, tagValue := range tags {
		insertTagsArgs.Append(tagValue)
		insertTagsFields = append(insertTagsFields, "("+insertTagsArgs.Placeholder+")")
	}

	insertTagsQuery := `
    INSERT INTO tags
      (value)
    VALUES ` + strings.Join(insertTagsFields, ",") + `
      ON CONFLICT (value) DO UPDATE
      SET value = tags.value RETURNING id`

	rows, err := tx.QueryxContext(ctx, insertTagsQuery, insertTagsArgs.Values...)
	if err != nil {
		return err
	}

	insertedTagIds := []uint64{}

	for rows.Next() {
		var tagId uint64
		if err := rows.Scan(&tagId); err != nil {
			return err
		}
		insertedTagIds = append(insertedTagIds, tagId)
	}
	if err := rows.Close(); err != nil {
		return err
	}

	articleRelFields := []string{}
	articleRelArgs := NewArgs()

	for _, tagId := range insertedTagIds {
		articleRelArgs.Append(tagId)
		field := "(" + articleRelArgs.Placeholder + ","
		articleRelArgs.Append(articleID)
		field += articleRelArgs.Placeholder + ")"
		articleRelFields = append(articleRelFields, field)
	}

	insertArticleRelQuery := `
    INSERT INTO tags_articles_rel
      (tag_id, article_id)
    VALUES ` + strings.Join(articleRelFields, ",")

	_, err = tx.ExecContext(ctx, insertArticleRelQuery, articleRelArgs.Values...)

	return err
}

func (s *Storage) SelectTags(ctx context.Context) ([]string, error) {
	const query = `SELECT value FROM tags`
	rows, err := s.db.QueryxContext(ctx, query)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []string{}, nil
		default:
			return nil, err
		}
	}

	tags := []string{}
	for rows.Next() {
		var tag string
		rows.Scan(&tag)
		tags = append(tags, tag)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return tags, nil
}
