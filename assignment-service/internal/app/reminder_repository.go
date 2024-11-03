package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"go.opentelemetry.io/otel/trace"
)

type ReminderRepository struct {
	db     *database.Service
	tracer trace.Tracer
}

func NewReminderRepository(db *database.Service, tracer trace.Tracer) *ReminderRepository {
	return &ReminderRepository{db, tracer}
}

func (r *ReminderRepository) InsertReminder(ctx context.Context, reminder Reminder) (Reminder, error) {
	ctx, span := r.tracer.Start(ctx, "ReminderRepository.InsertReminder")
	defer span.End()

	query := `
	INSERT INTO reminders (assignment_id, date) 
	VALUES ($1, $2)
	RETURNING id, assignment_id, date, created_at, updated_at
	`

	var rd Reminder
	err := r.db.Pool.QueryRowContext(ctx, query, reminder.AssignmentID, reminder.Date).Scan(
		&rd.ID, &rd.AssignmentID, &rd.Date, &rd.CreatedAt, &rd.UpdatedAt,
	)

	return rd, err
}

func (r *ReminderRepository) InsertRemindersWithTransaction(ctx context.Context, tx *sql.Tx, reminders []Reminder) ([]Reminder, error) {
	ctx, span := r.tracer.Start(ctx, "ReminderRepository.InsertRemindersWithTransaction")
	defer span.End()

	var insertedReminders []Reminder
	if len(reminders) == 0 {
		return insertedReminders, nil
	}

	var values []string
	var args []interface{}
	for i, rem := range reminders {
		values = append(values, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, rem.AssignmentID, rem.Date)
	}

	query := fmt.Sprintf(`
	INSERT INTO reminders (assignment_id, date) 
	VALUES %s 
	RETURNING id, assignment_id, date, created_at, updated_at`, strings.Join(values, ", "))

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return insertedReminders, err
	}
	defer rows.Close()

	for rows.Next() {
		var rem Reminder
		if err := rows.Scan(&rem.ID, &rem.AssignmentID, &rem.Date, &rem.CreatedAt, &rem.UpdatedAt); err != nil {
			return nil, err
		}
		insertedReminders = append(insertedReminders, rem)
	}

	return insertedReminders, nil
}

func (r *ReminderRepository) DeleteRemindersByAssignmentIDWithTransaction(ctx context.Context, tx *sql.Tx, assignmentID int) error {
	ctx, span := r.tracer.Start(ctx, "ReminderRepository.DeleteRemindersByAssignmentIDWithTransaction")
	defer span.End()

	query := `
	DELETE FROM reminders
	WHERE assignment_id = $1
	`

	_, err := tx.ExecContext(ctx, query, assignmentID)
	return err
}
