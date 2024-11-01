package app

import (
	"context"
	"database/sql"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
)

type MissionRepository struct {
	db *database.Service
}

func NewMissionRepository(db *database.Service) *MissionRepository {
	return &MissionRepository{db: db}
}

func (r *MissionRepository) GetUserMissions(ctx context.Context, userID string) ([]Mission, error) {
	missions := []Mission{}

	query := `
	SELECT m.id, m.title, m.illustration, m.goal, m.reward, m.created_at, m.updated_at
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
		err := rows.Scan(&m.ID, &m.Title, &m.Illustration, &m.Goal, &m.Reward, &m.CreatedAt, &m.UpdatedAt)
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
