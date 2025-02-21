package model

type RoundStats struct {
	ID                 uint `gorm:"primarykey"`
	FightStatsID       uint `gorm:"foreignKey:ID;references:FightStatsID"`
	Round              int
	TotalStrikesThrown int
	TotalStrikesLanded int
	SigStrikesThrown   int
	SigStrikesLanded   int
	TakedownsAttempted int
	TakedownsLanded    int
	Outcome            string // either "WIND","LOSED","WINF","LOSEF","DRAW","WINNC","LOSENC","WINDQ","LOSEDQ", or "RE"(round ended)
	TimeLeft           int    // time left when round ended in seconds (usually 0)
	Knockdowns         int
	ControlTime        int
	SubAttempts        int
}
