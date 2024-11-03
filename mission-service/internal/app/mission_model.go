package app

type Mission struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	Illustration     string `json:"illustration"`
	Goal             int    `json:"goal"`
	Reward           int    `json:"reward"`
	EventEncreasorID int    `json:"-"`
	EventDecreasorID int    `json:"-"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}
