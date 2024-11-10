package app

type Mission struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	ImagePath        string `json:"image_path"`
	Goal             int    `json:"goal"`
	Reward           int    `json:"reward"`
	EventEncreasorID int    `json:"-"`
	EventDecreasorID int    `json:"-"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}
