package app

import (
	"context"
	"database/sql"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
)

type AssignmentRepository struct {
	db *database.Service
}

func NewAssignmentRepository(db *database.Service) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) InsertAssignmentWithTransaction(ctx context.Context, tx *sql.Tx, assignment Assignment) (Assignment, error) {
	query := `
	INSERT INTO assignments (user_id, title, note, due_date) 
	VALUES ($1, $2, $3, $4)
	RETURNING id, user_id, title, note, due_date, created_at, updated_at
	`

	var a Assignment
	err := tx.QueryRowContext(ctx, query, assignment.UserID, assignment.Title, assignment.Note, assignment.DueDate).Scan(
		&a.ID, &a.UserID, &a.Title, &a.Note, &a.DueDate, &a.CreatedAt, &a.UpdatedAt,
	)

	return a, err

}

func (r *AssignmentRepository) FindAssignmentsByUserIDJoinReminders(ctx context.Context, userID string) ([]Assignment, error) {
	query := `
	SELECT a.id, a.user_id, a.title, a.note, a.due_date, a.is_completed, a.is_important, a.created_at, a.updated_at, r.id, r.assignment_id, r.date, r.created_at, r.updated_at
	FROM assignments a
	LEFT JOIN reminders r ON a.id = r.assignment_id
	WHERE a.user_id = $1
	`

	rows, err := r.db.Pool.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := make([]Assignment, 0)
	for rows.Next() {
		var a Assignment
		var r Reminder
		err := rows.Scan(&a.ID, &a.UserID, &a.Title, &a.Note, &a.DueDate, &a.IsCompleted, &a.IsImportant, &a.CreatedAt, &a.UpdatedAt,
			&r.ID, &r.AssignmentID, &r.Date, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}

		assignmentExist := false
		for i, assignment := range assignments {
			if assignment.ID == a.ID {
				assignments[i].Reminders = append(assignments[i].Reminders, r)
				assignmentExist = true
				break
			}
		}

		if !assignmentExist {
			a.Reminders = append(a.Reminders, r)
			assignments = append(assignments, a)
		}
	}

	return assignments, nil
}

func (r *AssignmentRepository) FindAssignmentByIDJoinReminders(ctx context.Context, assignmentID int) (Assignment, error) {
	query := `
	SELECT a.id, a.user_id, a.title, a.note, a.due_date, a.is_completed, a.is_important, a.created_at, a.updated_at, r.id, r.assignment_id, r.date, r.created_at, r.updated_at
	FROM assignments a
	LEFT JOIN reminders r ON a.id = r.assignment_id
	WHERE a.id = $1
	`

	var a Assignment
	row, err := r.db.Pool.QueryContext(ctx, query, assignmentID)
	if err != nil {
		return a, err
	}
	defer row.Close()

	rowFound := false
	for row.Next() {
		rowFound = true
		var r Reminder
		err = row.Scan(&a.ID, &a.UserID, &a.Title, &a.Note, &a.DueDate, &a.IsCompleted, &a.IsImportant, &a.CreatedAt, &a.UpdatedAt, &r.ID, &r.AssignmentID, &r.Date, &r.CreatedAt, &r.UpdatedAt)
		a.Reminders = append(a.Reminders, r)
	}
	if !rowFound {
		return a, sql.ErrNoRows
	}

	return a, err
}

func (r *AssignmentRepository) FindAssignmentByID(ctx context.Context, assignmentID int) (Assignment, error) {
	query := `
	SELECT id, user_id, title, note, due_date, is_completed, is_important, created_at, updated_at
	FROM assignments
	WHERE id = $1
	`

	var a Assignment
	err := r.db.Pool.QueryRowContext(ctx, query, assignmentID).Scan(
		&a.ID, &a.UserID, &a.Title, &a.Note, &a.DueDate, &a.IsCompleted, &a.IsImportant, &a.CreatedAt, &a.UpdatedAt,
	)

	return a, err
}

func (r *AssignmentRepository) DeleteAssignmentByIDWithTransaction(ctx context.Context, tx *sql.Tx, assignmentID int) error {
	query := `
	DELETE FROM assignments
	WHERE id = $1
	`

	_, err := tx.ExecContext(ctx, query, assignmentID)
	return err
}

func (r *AssignmentRepository) UpdateAssignmentWithTransaction(ctx context.Context, tx *sql.Tx, assignment Assignment) (Assignment, error) {
	query := `
	UPDATE assignments
	SET title = $1, note = $2, due_date = $3, is_completed = $4, is_important = $5, updated_at = NOW()
	WHERE id = $6
	RETURNING id, user_id, title, note, due_date, is_completed, is_important, created_at, updated_at
	`

	var a Assignment
	err := tx.QueryRowContext(ctx, query, assignment.Title, assignment.Note, assignment.DueDate, assignment.IsCompleted, assignment.IsImportant, assignment.ID).Scan(
		&a.ID, &a.UserID, &a.Title, &a.Note, &a.DueDate, &a.IsCompleted, &a.IsImportant, &a.CreatedAt, &a.UpdatedAt,
	)

	return a, err
}
