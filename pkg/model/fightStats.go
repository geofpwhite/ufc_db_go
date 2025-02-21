package model

type FightStats struct {
	ID                 uint `gorm:"primarykey"`
	FighterID          uint
	FightID            uint
	TotalStrikesThrown int
	TotalStrikesLanded int
	SigStrikesThrown   int
	SigStrikesLanded   int
	TakedownsAttempted int
	TakedownsLanded    int
	RequiredWeight     int
	Weight             int
	Outcome            string // either "WIND","LOSED","WINF","LOSEF","DRAW","WINNC","LOSENC","WINDQ","LOSEDQ"
	FinalRound         int    // the round in which the fight ended, so either the final round of the fight or the round when one of them got finished
	TimeLeft           int    // time left in final round in seconds
	Knockdowns         int
	ControlTime        int
	SubAttempts        int
	Referee            string
	RoundStats         []RoundStats `gorm:"foreignKey:FightStatsID;references:ID"`
}
