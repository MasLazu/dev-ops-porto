package app

import "time"

type Reminder struct {
	ID           int       `json:"id"`
	AssignmentID int       `json:"assignment_id"`
	Date         time.Time `json:"date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateReminderRequest struct {
	Date time.Time `json:"date" validate:"required,future"`
}
