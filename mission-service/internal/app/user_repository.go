package app

import (
	"context"
	"database/sql"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
)

type UserRepository struct {
	db *database.Service
}

func NewUserRepository(db *database.Service) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (User, error) {
	user := User{}

	query := `
	SELECT id, expiration_date, created_at, updated_at
	FROM users
	WHERE id = $1
	`

	err := r.db.Pool.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.ExpirationDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func (r *UserRepository) InsertUserWithTransaction(ctx context.Context, tx *sql.Tx, user User) (User, error) {
	query := `
	INSERT INTO users (id, expiration_date)
	VALUES ($1, $2)
	RETURNING id, expiration_date, created_at, updated_at
	`

	err := tx.QueryRowContext(ctx, query, user.ID, user.ExpirationDate).Scan(
		&user.ID,
		&user.ExpirationDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func (r *UserRepository) UpdateUserWithTransaction(ctx context.Context, tx *sql.Tx, user User) (User, error) {
	query := `
	UPDATE users
	SET expiration_date = $2, updated_at = now()
	WHERE id = $1
	RETURNING id, expiration_date, created_at, updated_at
	`

	err := tx.QueryRowContext(ctx, query, user.ID, user.ExpirationDate).Scan(
		&user.ID,
		&user.ExpirationDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}
