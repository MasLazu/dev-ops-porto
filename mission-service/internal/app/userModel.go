package app

import "time"

type User struct {
	ID             string    `json:"id"`
	ExpirationDate time.Time `json:"expiration_date"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

type UserExpirationMissionDateResponse struct {
	ExpirationDate time.Time `json:"expiration_date"`
}
