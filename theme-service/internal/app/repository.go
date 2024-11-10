package app

import (
	"context"
	"database/sql"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"go.opentelemetry.io/otel/trace"
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

func (r *Repository) FindOwnedThemes(ctx context.Context, userID string) ([]Theme, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.FindOwnedThemes")
	defer span.End()

	query := `
	SELECT t.id, t.name, t.price, t.created_at, t.updated_at
	FROM themes t
	JOIN owned_themes ot ON t.id = ot.theme_id
	WHERE ot.user_id = $1
	`
	rows, err := r.db.Pool.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var themes []Theme
	for rows.Next() {
		var t Theme
		err = rows.Scan(&t.ID, &t.Name, &t.Price, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}

	if len(themes) == 0 {
		return themes, sql.ErrNoRows
	}

	return themes, nil
}

func (r *Repository) FindThemeByID(ctx context.Context, themeID int) (Theme, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.FindThemeByID")
	defer span.End()

	query := `
	SELECT id, name, created_at, updated_at
	FROM themes
	WHERE id = $1
	`
	row := r.db.Pool.QueryRowContext(ctx, query, themeID)

	var t Theme
	err := row.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return Theme{}, err
	}

	return t, nil
}

func (r *Repository) InserOwnedTheme(ctx context.Context, userID string, themeID int) error {
	ctx, span := r.tracer.Start(ctx, "Repository.InsertOwnedTheme")
	defer span.End()

	query := `
	INSERT INTO owned_themes (user_id, theme_id)
	VALUES ($1, $2)
	`
	_, err := r.db.Pool.ExecContext(ctx, query, userID, themeID)
	if err != nil {
		return err
	}

	return nil
}
