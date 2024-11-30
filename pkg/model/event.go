package model

import (
	"time"
)

type Event struct {
	ID       uint `gorm:"primarykey"`
	Date     time.Time
	Title    string
	Fights   []Fight `gorm:"foreignKey:EventID;references:ID"`
	Location string
}
