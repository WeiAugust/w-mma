package ufc

import "time"

type EventLink struct {
	Name      string
	URL       string
	StartsAt  time.Time
	PosterURL string
}

type EventBout struct {
	RedName     string
	RedURL      string
	RedRank     string
	BlueName    string
	BlueURL     string
	BlueRank    string
	WeightClass string
	CardSegment string
	WinnerSide  string
	Result      string
	Method      string
	Round       int
	TimeSec     int
}

type EventCard struct {
	Name      string
	URL       string
	Status    string
	StartsAt  time.Time
	Venue     string
	PosterURL string
	Bouts     []EventBout
}

type AthleteProfile struct {
	Name      string
	URL       string
	Country   string
	Record    string
	AvatarURL string
}

type EventRecord struct {
	SourceID    int64
	Org         string
	Name        string
	Status      string
	StartsAt    time.Time
	Venue       string
	PosterURL   string
	ExternalURL string
}

type FighterRecord struct {
	SourceID    int64
	Name        string
	Country     string
	Record      string
	WeightClass string
	AvatarURL   string
	ExternalURL string
}

type BoutRecord struct {
	RedFighterID  int64
	BlueFighterID int64
	CardSegment   string
	WeightClass   string
	RedRanking    string
	BlueRanking   string
	Result        string
	WinnerID      int64
	Method        string
	Round         int
	TimeSec       int
}

type SyncResult struct {
	Events   int `json:"events"`
	Bouts    int `json:"bouts"`
	Fighters int `json:"fighters"`
}
