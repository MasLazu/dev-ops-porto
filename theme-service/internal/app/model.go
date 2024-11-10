package app

type Theme struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UnlockThemeRequest struct {
	ThemeID int `json:"theme_id" validate:"required"`
}
