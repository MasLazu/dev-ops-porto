package app

import (
	"context"
	"database/sql"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
)

type Repository struct {
	db     *database.Service
	tracer trace.Tracer
}

func NewRepository(db *database.Service, tracer trace.Tracer) *Repository {
	return &Repository{db, tracer}
}

func (r *Repository) Health(ctx context.Context) map[string]string {
	return r.db.Health(ctx)
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (user, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.FindUserByEmail")
	defer span.End()

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

func (r *Repository) FindUserByID(ctx context.Context, userID string) (user, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.FindUserByID")
	defer span.End()

	query := `
	SELECT id, email, name, coin, profile_picture, created_at, updated_at, password
	FROM users 
	WHERE id = $1
	`
	row := r.db.Pool.QueryRowContext(ctx, query, userID)

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
	ctx, span := r.tracer.Start(ctx, "Repository.InsertUser")
	defer span.End()

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
