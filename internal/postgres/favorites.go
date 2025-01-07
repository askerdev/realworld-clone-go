package postgres

import (
	"context"
	"database/sql"
)

func (s *Storage) FavoriteArticle(ctx context.Context, userID uint64, articleID uint64) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	const insertQuery = `
    INSERT INTO favorites_articles_rel
      (user_id, article_id)
    VALUES
      ($1, $2)`

	_, err = tx.ExecContext(ctx, insertQuery, userID, articleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	const incrementQuery = `UPDATE articles SET favorites_count = favorites_count + 1 WHERE id = $1`

	_, err = tx.ExecContext(ctx, incrementQuery, articleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) UnfavoriteArticle(ctx context.Context, userID uint64, articleID uint64) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	const deleteQuery = `DELETE FROM favorites_articles_rel WHERE user_id = $1 AND article_id = $2`

	_, err = tx.ExecContext(ctx, deleteQuery, userID, articleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	const decrementQuery = `UPDATE articles SET favorites_count = favorites_count - 1 WHERE id = $1`

	_, err = tx.ExecContext(ctx, decrementQuery, articleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
