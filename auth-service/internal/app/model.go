package app

import "time"

type user struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Coin           int       `json:"coin"`
	ProfilePicture string    `json:"profile_picture,omitempty"`
	Password       string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type registerUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (rur *registerUserRequest) toUser() user {
	return user{
		Email:    rur.Email,
		Name:     rur.Name,
		Password: rur.Password,
	}
}

type loginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}
