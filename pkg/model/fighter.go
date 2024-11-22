package model

import (
	"time"

	"gorm.io/gorm"
)

type Fighter struct {
	gorm.Model
	ID         uint `gorm:"primarykey"`
	Name       string
	DOB        time.Time
	FightStats []FightStats `gorm:"foreignKey:FighterID;references:ID"`
}
