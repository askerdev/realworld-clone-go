package postgres

import (
	"context"
	"strconv"
	"strings"
)

func (s *Storage) InsertTags(ctx context.Context, articleID uint64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	count := 1
	fields := []string{}
	args := []any{}

	for _, tag := range tags {
		args = append(args, tag, articleID)
		fields = append(fields, "($"+strconv.Itoa(count)+", $"+strconv.Itoa(count+1)+")")
		count += 2
	}

	insertTagsQuery := `
    INSERT INTO tags 
      (value, article_id)
    VALUES ` + strings.Join(fields, ",")

	_, err := s.db.ExecContext(ctx, insertTagsQuery, args...)
	return err
}
