package app

type Mission struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Illustration string `json:"illustration"`
	Goal         string `json:"goal"`
	Reward       string `json:"reward"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
