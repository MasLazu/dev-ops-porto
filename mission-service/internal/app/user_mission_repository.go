package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"go.opentelemetry.io/otel/trace"
)

type UserMissionRepository struct {
	db     *database.Service
	tracer trace.Tracer
}

func NewUserMissionRepository(db *database.Service, tracer trace.Tracer) *UserMissionRepository {
	return &UserMissionRepository{db, tracer}
}

func (r *UserMissionRepository) DeleteUserMissionByUserIDWithTransaction(
	ctx context.Context,
	tx *sql.Tx,
	userID string,
) error {
	ctx, span := r.tracer.Start(ctx, "UserMissionRepository.DeleteUserMissionByUserIDWithTransaction")
	defer span.End()

	query := `
	DELETE FROM users_missions
	WHERE user_id = $1
	`

	_, err := tx.ExecContext(ctx, query, userID)

	return err
}

func (r *UserMissionRepository) InsertUserMissionsWithTransaction(
	ctx context.Context,
	tx *sql.Tx,
	userID string,
	missionIDs []int,
) error {
	ctx, span := r.tracer.Start(ctx, "UserMissionRepository.InsertUserMissionsWithTransaction")
	defer span.End()

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

func (r *UserMissionRepository) UpdateUserMissionWithTransaction(
	ctx context.Context,
	tx *sql.Tx,
	userMission UserMission,
) error {
	ctx, span := r.tracer.Start(ctx, "UserMissionRepository.UpdateUserMission")
	defer span.End()

	query := `
	UPDATE users_missions
	SET progress = $1, claimed = $2
	WHERE id = $3
	`

	_, err := tx.ExecContext(ctx, query, userMission.Progress, userMission.Claimed, userMission.ID)

	return err
}

func (r *UserMissionRepository) UpdateUserMissions(
	ctx context.Context,
	missions []UserMission,
) error {
	ctx, span := r.tracer.Start(ctx, "UserMissionRepository.UpdateUserMissionsWithTransaction")
	defer span.End()

	if len(missions) == 0 {
		return nil
	}

	progressCaseStatements := ``
	claimedCaseStatements := ``
	whereStatements := ``
	args := make([]interface{}, len(missions)*3)
	for i, um := range missions {
		progressCaseStatements += fmt.Sprintf("WHEN id = $%d::integer THEN $%d::integer ", i*3+1, i*3+2)
		claimedCaseStatements += fmt.Sprintf("WHEN id = $%d::integer THEN $%d::boolean ", i*3+1, i*3+3)
		if i == len(missions)-1 {
			whereStatements += fmt.Sprintf("$%d::integer", i*3+1)
		} else {
			whereStatements += fmt.Sprintf("$%d::integer, ", i*3+1)
		}
		args[i*3] = um.ID
		args[i*3+1] = um.Progress
		args[i*3+2] = um.Claimed
	}

	query := fmt.Sprintf(`
	UPDATE users_missions
	SET 
		progress = CASE %sEND,
		claimed = CASE %sEND
	WHERE id IN (%s)
	`, progressCaseStatements, claimedCaseStatements, whereStatements)

	log.Printf("args: %v", args)
	_, err := r.db.Pool.ExecContext(ctx, query, args...)

	return err
}

func (r *UserMissionRepository) GetUserMissionByIDJoinMission(ctx context.Context, userMissionID int) (UserMission, error) {
	ctx, span := r.tracer.Start(ctx, "UserMissionRepository.GetUserMissionByID")
	defer span.End()

	query := `
	SELECT um.id, um.user_id, um.mission_id, um.progress, um.claimed, um.created_at, um.updated_at,
	m.id, m.title, m.illustration, m.event_encreasor_id, m.event_decreasor_id, m.goal, m.reward, m.created_at, m.updated_at
	FROM users_missions um
	JOIN missions m ON m.id = um.mission_id
	WHERE um.id = $1
	`

	var um UserMission
	err := r.db.Pool.QueryRowContext(ctx, query, userMissionID).
		Scan(&um.ID, &um.UserID, &um.MissionID, &um.Progress, &um.Claimed, &um.CreatedAt, &um.UpdatedAt,
			&um.Mission.ID, &um.Mission.Title, &um.Mission.Illustration, &um.Mission.EventEncreasorID,
			&um.Mission.EventDecreasorID, &um.Mission.Goal, &um.Mission.Reward, &um.Mission.CreatedAt, &um.Mission.UpdatedAt)

	return um, err
}

func (r *UserMissionRepository) GetUserMissionsByUserIDJoinMission(ctx context.Context, userID string) ([]UserMission, error) {
	ctx, span := r.tracer.Start(ctx, "UserMissionRepository.GetUserMissionsByUserIDJoinMission")
	defer span.End()

	query := `
	SELECT um.id, um.user_id, um.mission_id, um.progress, um.claimed, um.created_at, um.updated_at,
	m.id, m.title, m.illustration, m.event_encreasor_id, m.event_decreasor_id, m.goal, m.reward, m.created_at, m.updated_at
	FROM users_missions um
	JOIN missions m ON m.id = um.mission_id
	WHERE um.user_id = $1
	`

	userMissions := make([]UserMission, 0)

	rows, err := r.db.Pool.QueryContext(ctx, query, userID)
	if err != nil {
		return userMissions, err
	}
	defer rows.Close()

	for rows.Next() {
		var um UserMission
		var eventDecreasorID sql.NullInt32
		err := rows.Scan(&um.ID, &um.UserID, &um.MissionID, &um.Progress, &um.Claimed, &um.CreatedAt, &um.UpdatedAt,
			&um.Mission.ID, &um.Mission.Title, &um.Mission.Illustration, &um.Mission.EventEncreasorID,
			&eventDecreasorID, &um.Mission.Goal, &um.Mission.Reward, &um.Mission.CreatedAt, &um.Mission.UpdatedAt)
		if err != nil {
			return userMissions, err
		}

		if eventDecreasorID.Valid {
			um.Mission.EventDecreasorID = int(eventDecreasorID.Int32)
		}

		userMissions = append(userMissions, um)
	}

	return userMissions, nil
}

func (r *UserMissionRepository) GetUserMissionsByUserIDAndEncreasorEventIDJoinMission(
	ctx context.Context,
	userID string,
	eventID int,
) ([]UserMission, error) {
	ctx, span := r.tracer.Start(ctx, "UserMissionRepository.GetUserMissionsByUserIDAndEncreasorEventIDJoinMission")
	defer span.End()

	missions := []UserMission{}

	query := `
	SELECT um.id, um.user_id, um.mission_id, um.progress, um.claimed, um.created_at, um.updated_at,
	m.id, m.title, m.illustration, m.event_encreasor_id, m.event_decreasor_id, m.goal, m.reward, m.created_at, m.updated_at
	FROM users_missions um
	JOIN missions m ON m.id = um.mission_id
	WHERE um.user_id = $1 AND m.event_encreasor_id = $2
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, userID, eventID)
	if err != nil {
		return missions, err
	}
	defer rows.Close()

	empty := true
	for rows.Next() {
		var um UserMission
		empty = false
		err := rows.Scan(&um.ID, &um.UserID, &um.MissionID, &um.Progress, &um.Claimed, &um.CreatedAt, &um.UpdatedAt,
			&um.Mission.ID, &um.Mission.Title, &um.Mission.Illustration, &um.Mission.EventEncreasorID,
			&um.Mission.EventDecreasorID, &um.Mission.Goal, &um.Mission.Reward, &um.Mission.CreatedAt, &um.Mission.UpdatedAt)
		if err != nil {
			return missions, err
		}

		missions = append(missions, um)
	}
	if empty {
		return missions, sql.ErrNoRows
	}

	return missions, nil
}

func (r *UserMissionRepository) GetUserMissionsByUserIDAndDecreasorEventIDJoinMission(
	ctx context.Context,
	userID string,
	eventID int,
) ([]UserMission, error) {
	ctx, span := r.tracer.Start(ctx, "UserMissionRepository.GetUserMissionsByUserIDAndDecreasorEventIDJoinMission")
	defer span.End()

	missions := []UserMission{}

	query := `
	SELECT um.id, um.user_id, um.mission_id, um.progress, um.claimed, um.created_at, um.updated_at, 
	m.id, m.title, m.illustration, m.event_encreasor_id, m.event_decreasor_id, m.goal, m.reward, m.created_at, m.updated_at
	FROM users_missions um
	JOIN missions m ON m.id = um.mission_id
	WHERE um.user_id = $1 AND m.event_decreasor_id = $2
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, userID, eventID)
	if err != nil {
		return missions, err
	}
	defer rows.Close()

	empty := true
	for rows.Next() {
		var um UserMission
		var eventDecreasorID sql.NullInt32
		empty = false
		err := rows.Scan(&um.ID, &um.UserID, &um.MissionID, &um.Progress, &um.Claimed, &um.CreatedAt, &um.UpdatedAt,
			&um.Mission.ID, &um.Mission.Title, &um.Mission.Illustration, &um.Mission.EventEncreasorID,
			&eventDecreasorID, &um.Mission.Goal, &um.Mission.Reward, &um.Mission.CreatedAt, &um.Mission.UpdatedAt)
		if err != nil {
			return missions, err
		}

		if eventDecreasorID.Valid {
			um.Mission.EventDecreasorID = int(eventDecreasorID.Int32)
		}

		missions = append(missions, um)
	}
	if empty {
		return missions, sql.ErrNoRows
	}

	return missions, nil
}
