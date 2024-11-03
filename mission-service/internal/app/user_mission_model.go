package app

type UserMission struct {
	ID        int
	UserID    string
	MissionID int
	Progress  int
	Claimed   bool
	CreatedAt string
	UpdatedAt string
	Mission   Mission
}
