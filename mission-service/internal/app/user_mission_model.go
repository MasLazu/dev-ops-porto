package app

type UserMission struct {
	ID        int     `json:"id"`
	UserID    string  `json:"user_id"`
	MissionID int     `json:"mission_id"`
	Progress  int     `json:"progress"`
	Claimed   bool    `json:"claimed"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	Mission   Mission `json:"mission"`
}
