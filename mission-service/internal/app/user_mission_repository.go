package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
)

type UserMissionRepository struct {
	db *database.Service
}

func NewUserMissionRepository(db *database.Service) *UserMissionRepository {
	return &UserMissionRepository{db: db}
}

func (r *UserMissionRepository) DeleteUserMissionByUserIDWithTransaction(ctx context.Context, tx *sql.Tx, userID string) error {
	query := `
	DELETE FROM users_missions
	WHERE user_id = $1
	`

	_, err := tx.ExecContext(ctx, query, userID)

	return err
}

func (r *UserMissionRepository) InsertUserMissionsWithTransaction(ctx context.Context, tx *sql.Tx, userID string, missionIDs []int) error {
	var values []string
	args := make([]interface{}, len(missionIDs)+1)
	args[0] = userID
	for i := range missionIDs {
		values = append(values, fmt.Sprintf("($1, $%d)", i+2))
		args[i+1] = missionIDs[i]
	}

	query := fmt.Sprintf(`
	INSERT INTO users_missions (user_id, mission_id)
	VALUES %s`, strings.Join(values, ", "))

	_, err := tx.ExecContext(ctx, query, args...)

	return err
}
