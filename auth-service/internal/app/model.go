package app

import "time"

type user struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Coin           int       `json:"coin"`
	ProfilePicture *string   `json:"profile_picture,omitempty"`
	Password       string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (u *user) addPrefixToProfilePictureURL(prefix string) {
	if u.ProfilePicture == nil {
		return
	}
	*u.ProfilePicture = prefix + *u.ProfilePicture
}

type RegisterUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (rur *RegisterUserRequest) toUser() user {
	return user{
		Email:    rur.Email,
		Name:     rur.Name,
		Password: rur.Password,
	}
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
