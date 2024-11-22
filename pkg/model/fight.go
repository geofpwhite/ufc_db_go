package model

import "gorm.io/gorm"

type Fight struct {
	gorm.Model
	ID            uint `gorm:"primarykey"`
	EventID       uint
	Fighter1Stats FightStats
	Fighter2Stats FightStats
}
