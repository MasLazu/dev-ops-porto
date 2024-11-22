package app

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/MasLazu/dev-ops-porto/assignment-service/.gen/database/public/table"
	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	//lint:ignore ST1001 "github.com/go-jet/jet/v2/postgres"
	. "github.com/go-jet/jet/v2/postgres"
)

var AssignmentTable = table.Assignments.AS("assignment")

type AssignmentRepository struct {
	db     *database.Service
	tracer trace.Tracer
}

func NewAssignmentRepository(db *database.Service, tracer trace.Tracer) *AssignmentRepository {
	return &AssignmentRepository{db: db, tracer: tracer}
}

func (r *AssignmentRepository) InsertAssignmentWithTransaction(ctx context.Context, tx *sql.Tx, assignment Assignment) (Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "AssignmentRepository.InsertAssignmentWithTransaction")
	defer span.End()

	query := AssignmentTable.INSERT(AssignmentTable.UserID, AssignmentTable.Title, AssignmentTable.Note, AssignmentTable.DueDate).
		VALUES(assignment.UserID, assignment.Title, assignment.Note, assignment.DueDate).
		RETURNING(AssignmentTable.AllColumns)

	var a Assignment
	err := query.QueryContext(ctx, tx, &a)

	log.Println("assignment", a)

	return a, err
}

func (r *AssignmentRepository) FindAssignmentsByUserIDJoinReminders(ctx context.Context, userID uuid.UUID) ([]Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "AssignmentRepository.FindAssignmentsByUserIDJoinReminders")
	defer span.End()

	query := AssignmentTable.SELECT(AssignmentTable.AllColumns, ReminderTable.AllColumns).
		FROM(AssignmentTable.LEFT_JOIN(ReminderTable, AssignmentTable.ID.EQ(ReminderTable.AssignmentID))).
		WHERE(AssignmentTable.UserID.EQ(UUID(userID)))

	var AssignmentTable []Assignment
	err := query.QueryContext(ctx, r.db.Pool, &AssignmentTable)

	return AssignmentTable, err
}

func (r *AssignmentRepository) FindAssignmentByIDJoinReminders(ctx context.Context, assignmentID int32) (Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "AssignmentRepository.FindAssignmentByIDJoinReminders")
	defer span.End()

	query := AssignmentTable.SELECT(AssignmentTable.AllColumns, ReminderTable.AllColumns).
		FROM(AssignmentTable.LEFT_JOIN(ReminderTable, AssignmentTable.ID.EQ(ReminderTable.AssignmentID))).
		WHERE(AssignmentTable.ID.EQ(Int32(assignmentID)))

	var a Assignment
	err := query.QueryContext(ctx, r.db.Pool, &a)

	return a, err
}

func (r *AssignmentRepository) FindAssignmentByID(ctx context.Context, assignmentID int32) (Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "AssignmentRepository.FindAssignmentByID")
	defer span.End()

	query := AssignmentTable.SELECT(AssignmentTable.AllColumns).
		FROM(AssignmentTable).
		WHERE(AssignmentTable.ID.EQ(Int32(assignmentID)))

	var a Assignment
	err := query.QueryContext(ctx, r.db.Pool, &a)

	return a, err
}

func (r *AssignmentRepository) DeleteAssignmentByIDWithTransaction(ctx context.Context, tx *sql.Tx, assignmentID int32) error {
	ctx, span := r.tracer.Start(ctx, "AssignmentRepository.DeleteAssignmentByIDWithTransaction")
	defer span.End()

	query := AssignmentTable.DELETE().
		WHERE(AssignmentTable.ID.EQ(Int32(assignmentID)))

	_, err := query.ExecContext(ctx, tx)
	return err
}

func (r *AssignmentRepository) UpdateAssignmentWithTransaction(ctx context.Context, tx *sql.Tx, assignment Assignment) (Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "AssignmentRepository.UpdateAssignmentWithTransaction")
	defer span.End()

	query := AssignmentTable.UPDATE(AssignmentTable.Title, AssignmentTable.Note, AssignmentTable.DueDate, AssignmentTable.IsCompleted, AssignmentTable.IsImportant, AssignmentTable.UpdatedAt).
		SET(assignment.Title, assignment.Note, assignment.DueDate, assignment.IsCompleted, assignment.IsImportant, time.Now()).
		WHERE(AssignmentTable.ID.EQ(Int32(assignment.ID))).
		RETURNING(AssignmentTable.AllColumns)

	var a Assignment
	err := query.QueryContext(ctx, tx, &a)

	return a, err
}
