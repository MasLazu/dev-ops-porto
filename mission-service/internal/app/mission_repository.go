package app

import (
	"context"
	"database/sql"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"go.opentelemetry.io/otel/trace"
)

type MissionRepository struct {
	db     *database.Service
	tracer trace.Tracer
}

func NewMissionRepository(db *database.Service, tracer trace.Tracer) *MissionRepository {
	return &MissionRepository{db, tracer}
}

func (r *MissionRepository) GetUserMissions(ctx context.Context, userID string) ([]Mission, error) {
	ctx, span := r.tracer.Start(ctx, "MissionRepository.GetUserMissions")
	defer span.End()

	missions := []Mission{}

	query := `
	SELECT m.id, m.title, m.image_path, m.goal, m.reward, m.created_at, m.updated_at
	FROM missions m
	JOIN users_missions um ON um.mission_id = m.id
	JOIN users u ON u.id = um.user_id
	WHERE u.id = $1
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, userID)
	if err != nil {
		return missions, err
	}
	defer rows.Close()

	empty := true
	for rows.Next() {
		var m Mission
		empty = false
		err := rows.Scan(&m.ID, &m.Title, &m.ImagePath, &m.Goal, &m.Reward, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return missions, err
		}

		missions = append(missions, m)
	}
	if empty {
		return missions, sql.ErrNoRows
	}

	return missions, nil
}

func (r *MissionRepository) GetTwoRandomMissionIDs(ctx context.Context) ([]int, error) {
	ctx, span := r.tracer.Start(ctx, "MissionRepository.GetTwoRandomMissionIDs")
	defer span.End()

	ids := []int{}

	query := `
	SELECT id
	FROM missions
	ORDER BY random()
	LIMIT 2
	`

	rows, err := r.db.Pool.QueryContext(ctx, query)
	if err != nil {
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
