package model

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	ID     uint `gorm:"primarykey"`
	Date   time.Time
	Title  string
	Fights []Fight `gorm:"foreignKey:EventID;references:ID"`
}
