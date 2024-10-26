package app

import (
	"assignment-service/internal/database"
	"context"
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
