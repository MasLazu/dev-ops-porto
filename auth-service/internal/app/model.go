package app

type user struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Coin           int    `json:"coin"`
	ProfilePicture string `json:"profile_picture,omitempty"`
	Password       string `json:"-"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
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
