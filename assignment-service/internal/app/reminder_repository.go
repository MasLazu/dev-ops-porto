package app

import (
	"context"
	"database/sql"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"go.opentelemetry.io/otel/trace"

	"github.com/MasLazu/dev-ops-porto/assignment-service/.gen/database/public/table"
	//lint:ignore ST1001 "github.com/go-jet/jet/v2/postgres"
	. "github.com/go-jet/jet/v2/postgres"
)

var ReminderTable = table.Reminders.AS("reminder")

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

	query := ReminderTable.INSERT(ReminderTable.AssignmentID, ReminderTable.Date).
		VALUES(reminder.AssignmentID, reminder.Date).
		RETURNING(ReminderTable.AllColumns)

	var rd Reminder
	err := query.QueryContext(ctx, r.db.Pool, &rd)

	return rd, err
}

func (r *ReminderRepository) InsertRemindersWithTransaction(ctx context.Context, tx *sql.Tx, reminders []Reminder) ([]Reminder, error) {
	ctx, span := r.tracer.Start(ctx, "ReminderRepository.InsertRemindersWithTransaction")
	defer span.End()

	query := ReminderTable.INSERT(ReminderTable.AssignmentID, ReminderTable.Date).
		RETURNING(ReminderTable.AllColumns)

	for _, rem := range reminders {
		query = query.VALUES(rem.AssignmentID, rem.Date)
	}

	var insertedReminders []Reminder
	err := query.QueryContext(ctx, tx, &insertedReminders)

	return insertedReminders, err
}

func (r *ReminderRepository) DeleteRemindersByAssignmentIDWithTransaction(ctx context.Context, tx *sql.Tx, assignmentID int32) error {
	ctx, span := r.tracer.Start(ctx, "ReminderRepository.DeleteRemindersByAssignmentIDWithTransaction")
	defer span.End()

	query := ReminderTable.DELETE().
		WHERE(ReminderTable.AssignmentID.EQ(Int32(assignmentID)))

	_, err := query.ExecContext(ctx, tx)

	return err
}
