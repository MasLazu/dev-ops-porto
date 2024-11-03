package app

type Mission struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	Illustration     string `json:"illustration"`
	Goal             int    `json:"goal"`
	Reward           int    `json:"reward"`
	EventEncreasorID int    `json:"event_encreasor_id"`
	EventDecreasorID int    `json:"event_decreasor_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}
