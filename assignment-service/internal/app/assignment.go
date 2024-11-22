package app

import (
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID          int32      `json:"id" sql:"primary_key"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	Note        *string    `json:"note"`
	DueDate     *time.Time `json:"due_date"`
	IsCompleted *bool      `json:"is_completed"`
	IsImportant *bool      `json:"is_important"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	Reminders   []Reminder `json:"reminders,omitempty"`
}

type CreateAssignmentRequest struct {
	Title       string                  `json:"title" validate:"required"`
	Note        *string                 `json:"note" validate:"required"`
	DueDate     *time.Time              `json:"due_date" validate:"required"`
	IsImportant *bool                   `json:"is_important"`
	Reminders   []CreateReminderRequest `json:"reminders,omitempty" validate:"dive"`
}

func (car *CreateAssignmentRequest) toAssignmentAndReminders(userID uuid.UUID) (Assignment, []Reminder) {
	var reminders []Reminder
	for _, reminder := range car.Reminders {
		reminders = append(reminders, Reminder{
			Date: reminder.Date,
		})
	}

	return Assignment{
		UserID:      userID,
		Title:       car.Title,
		Note:        car.Note,
		DueDate:     car.DueDate,
		IsImportant: car.IsImportant,
	}, reminders
}

type UpdateAssignmentRequest struct {
	Title       string                  `json:"title" validate:"required"`
	Note        *string                 `json:"note" validate:"required"`
	DueDate     *time.Time              `json:"due_date" validate:"required,future"`
	IsImportant *bool                   `json:"is_important"`
	IsCompleted *bool                   `json:"is_completed"`
	Reminders   []CreateReminderRequest `json:"reminders,omitempty" validate:"dive"`
}

type ChangeIsCompletedRequest struct {
	ID          int32 `json:"id" validate:"required"`
	IsCompleted bool  `json:"is_completed"`
}

func (uar *UpdateAssignmentRequest) toAssignment() Assignment {
	return Assignment{
		Title:       uar.Title,
		Note:        uar.Note,
		DueDate:     uar.DueDate,
		IsImportant: uar.IsImportant,
		IsCompleted: uar.IsCompleted,
	}
}
