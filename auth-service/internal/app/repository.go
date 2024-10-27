package app

import (
	"auth-service/internal/database"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository struct {
	db *database.Service
}

func NewRepository(db *database.Service) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Health(ctx context.Context) map[string]string {
	return r.db.Health(ctx)
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (user, error) {
	query := `
    SELECT id, email, name, coin, profile_picture, created_at, updated_at, password
    FROM users 
    WHERE email = $1
    `
	row := r.db.Pool.QueryRowContext(ctx, query, email)

	var u user
	var profilePicture sql.NullString
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Coin, &profilePicture, &u.CreatedAt, &u.UpdatedAt, &u.Password)
	if err != nil {
		return u, err
	}

	if profilePicture.Valid {
		u.ProfilePicture = profilePicture.String
	}

	return u, nil
}

func (r *Repository) InsertUser(ctx context.Context, u user) (user, error) {
	query := `
    INSERT INTO users (id, email, name, password)
    VALUES ($1, $2, $3, $4)
    RETURNING id, email, name, coin, created_at, updated_at
    `

	orderID := uuid.New().String()
	err := r.db.Pool.QueryRowContext(ctx, query, orderID, u.Email, u.Name, u.Password).Scan(
		&u.ID, &u.Email, &u.Name, &u.Coin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return u, err
	}

	return u, nil
}
