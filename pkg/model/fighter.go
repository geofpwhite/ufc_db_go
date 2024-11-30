package model

import (
	"time"
)

type Fighter struct {
	ID         uint `gorm:"primarykey"`
	FirstName  string
	LastName   string
	NickName   *string
	DOB        time.Time
	Height     *int
	Reach      *int
	FightStats []FightStats `gorm:"foreignKey:FighterID;references:ID"`
}
