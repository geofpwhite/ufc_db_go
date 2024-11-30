package model

type Fight struct {
	ID            uint `gorm:"primarykey"`
	EventID       uint
	Fighter1Stats FightStats
	Fighter2Stats FightStats
}
