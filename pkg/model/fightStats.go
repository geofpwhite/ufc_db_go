package model

import "gorm.io/gorm"

type FightStats struct {
	gorm.Model
	ID                 uint `gorm:"primarykey"`
	FighterID          uint
	FightID            uint
	TotalStrikesThrown int
	TotalStrikesLanded int
	SigStrikesThrown   int
	SigStrikesLanded   int
	TakedownsAttempted int
	TakedownsLanded    int
	Outcome            string // either "WIND","LOSED","WINF","LOSEF","DRAW","WINNC","LOSENC","WINDQ","LOSEDQ"
	FinalRound         int    // the round in which the fight ended, so either the final round of the fight or the round when one of them got finished
}
