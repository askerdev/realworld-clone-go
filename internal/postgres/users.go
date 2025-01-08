package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/askerdev/realworld-clone-go/internal/domain/entity"
	"github.com/guregu/null/v5"
	"github.com/jmoiron/sqlx"
)

func (r *Storage) InsertUser(
	ctx context.Context,
	email, username, password string,
) (*entity.User, error) {
	const query = `
    INSERT INTO users (email, username, password)
    VALUES ($1, $2, $3)
    RETURNING id, email, username, bio, image
    `
	row := r.db.QueryRowxContext(ctx, query, email, username, password)

	u := &entity.User{}
	if err := row.StructScan(u); err != nil {
		return nil, ErrUniqueConstraint
	}

	return u, nil
}

type SubscriptionRow struct {
	UserID    uint64 `db:"user_id"`
	ProfileID uint64 `db:"profile_id"`
}

func (r *Storage) FollowProfile(
	ctx context.Context,
	userId uint64,
	username string,
) (*entity.Profile, error) {
	const createSubscriptionQuery = `
    INSERT INTO subscriptions
      (user_id, profile_id)
    VALUES
      ($1, $2)`
	const selectProfileByUsernameQuery = `SELECT id, username, image, bio FROM users WHERE username = $1`
	const selectSubscription = `SELECT * FROM subscriptions WHERE user_id = $1 AND profile_id = $2`

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	p := &entity.Profile{}
	row := tx.QueryRowxContext(ctx, selectProfileByUsernameQuery, username)
	if err := row.StructScan(p); err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.ExecContext(ctx, createSubscriptionQuery, userId, p.ID)
	if err != nil {
		tx.Rollback()
		return nil, ErrSubscriptionAlreadyExists
	}

	sub := &SubscriptionRow{}
	row = tx.QueryRowxContext(ctx, selectSubscription, userId, p.ID)
	if err := row.StructScan(sub); err != nil {
		return nil, err
	}

	p.Following = sub.UserID == userId

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *Storage) UnfollowProfile(
	ctx context.Context,
	userId uint64,
	username string,
) (*entity.Profile, error) {
	const removeSubscriptionsQuery = `
    DELETE FROM subscriptions
    WHERE user_id = $1 AND profile_id = $2`
	const selectProfileByUsernameQuery = `SELECT id, username, image, bio FROM users WHERE username = $1`

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	p := &entity.Profile{}
	row := tx.QueryRowxContext(ctx, selectProfileByUsernameQuery, username)
	if err := row.StructScan(p); err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.ExecContext(ctx, removeSubscriptionsQuery, userId, p.ID)
	if err != nil {
		tx.Rollback()
		return nil, ErrSubscriptionAlreadyExists
	}

	p.Following = false

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *Storage) selectProfileByID(
	ctx context.Context,
	tx *sqlx.Tx,
	profileID uint64,
	userID *uint64,
) (*entity.Profile, error) {
	const selectProfileByIDQuery = `SELECT id, username, image, bio FROM users WHERE id = $1`

	p := &entity.Profile{}
	row := tx.QueryRowxContext(ctx, selectProfileByIDQuery, profileID)
	if err := row.StructScan(p); err != nil {
		return nil, err
	}

	if userID != nil {
		const selectSubscription = `SELECT * FROM subscriptions WHERE user_id = $1 AND profile_id = $2`
		sub := &SubscriptionRow{}
		row = tx.QueryRowxContext(ctx, selectSubscription, *userID, profileID)
		if err := row.StructScan(sub); err == nil {
			p.Following = sub.UserID == *userID
		}
	}

	return p, nil
}

func (r *Storage) SelectProfileByID(
	ctx context.Context,
	profileID uint64,
	userID *uint64,
) (*entity.Profile, error) {
	const selectProfileByIDQuery = `SELECT id, username, image, bio FROM users WHERE id = $1`

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	p := &entity.Profile{}
	row := tx.QueryRowxContext(ctx, selectProfileByIDQuery, profileID)
	if err := row.StructScan(p); err != nil {
		tx.Rollback()
		return nil, err
	}

	if userID != nil {
		const selectSubscription = `SELECT * FROM subscriptions WHERE user_id = $1 AND profile_id = $2`
		sub := &SubscriptionRow{}
		row = tx.QueryRowxContext(ctx, selectSubscription, *userID, profileID)
		if err := row.StructScan(sub); err == nil {
			p.Following = sub.UserID == *userID
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *Storage) SelectProfileByUsername(
	ctx context.Context,
	username string,
	userID *uint64,
) (*entity.Profile, error) {
	const selectProfileByUsernameQuery = `SELECT id, username, image, bio FROM users WHERE username = $1`
	const selectSubscription = `SELECT * FROM subscriptions WHERE user_id = $1 AND profile_id = $2`

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	p := &entity.Profile{}
	row := tx.QueryRowxContext(ctx, selectProfileByUsernameQuery, username)
	if err := row.StructScan(p); err != nil {
		tx.Rollback()
		return nil, err
	}

	if userID != nil {
		sub := &SubscriptionRow{}
		row = tx.QueryRowxContext(ctx, selectSubscription, *userID, p.ID)
		if err := row.StructScan(sub); err == nil {
			p.Following = sub.UserID == *userID
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *Storage) SelectUserByEmail(
	ctx context.Context,
	email string,
) (*entity.User, error) {
	const query = `SELECT * FROM users WHERE email = $1`
	row := r.db.QueryRowxContext(ctx, query, email)
	if row.Err() != nil {
		return nil, row.Err()
	}
	u := &entity.User{}
	if err := row.StructScan(u); err != nil {
		return nil, err
	}

	return u, nil
}

type UpdateUserParams struct {
	ID       uint64
	Email    null.String `db:"email"`
	Username null.String `db:"username"`
	Password null.String `db:"password"`
	Image    null.String `db:"image"`
	Bio      null.String `db:"bio"`
}

func (r *Storage) UpdateUser(
	ctx context.Context,
	updateUserParams *UpdateUserParams,
) (*entity.User, error) {
	fields := []string{}

	if updateUserParams.Email.Valid {
		fields = append(fields, "email = :email")
	}

	if updateUserParams.Username.Valid {
		fields = append(fields, "username = :username")
	}

	if updateUserParams.Password.Valid {
		fields = append(fields, "password = :password")
	}

	if updateUserParams.Image.Valid {
		fields = append(fields, "image = :image")
	}

	if updateUserParams.Bio.Valid {
		fields = append(fields, "bio = :bio")
	}

	query := `
    UPDATE users SET ` +
		strings.Join(fields, ",") +
		` WHERE id = :id
      RETURNING id, email, username, bio, image`
	rows, err := r.db.NamedQueryContext(ctx, query, updateUserParams)
	if err != nil {
		return nil, err
	}

	u := &entity.User{}
	for rows.Next() {
		if err := rows.StructScan(u); err != nil {
			return nil, err
		}
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return u, nil
}
