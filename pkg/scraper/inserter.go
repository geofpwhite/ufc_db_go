package scraper

import (
	"fmt"

	"github.com/geofpwhite/ufc_db_go/pkg/database"
	"github.com/geofpwhite/ufc_db_go/pkg/logger"
	"github.com/geofpwhite/ufc_db_go/pkg/model"
)

type Inserter struct {
	db database.Database
}

func NewInserter() *Inserter {
	return &Inserter{
		db: *database.Init(),
	}
}

func (i *Inserter) InsertEvents(eventChannel <-chan model.Event) {
	for event := range eventChannel {
		i.db.CreateEvent(event)
		logger.Logger.Printf("Inserted event {title:%s, date: %s", event.Title, event.Date)
	}
}

func (i *Inserter) InsertFighters(fighterChannel <-chan model.Fighter) {
	for fighter := range fighterChannel {
		i.db.CreateFighter(fighter)
		logger.Logger.Printf("Inserted fighter {name:%s %s}", fighter.FirstName, fighter.LastName)
	}
}

func (i *Inserter) InsertFights(fightChannel <-chan model.Fight) {
	for fight := range fightChannel {
		fmt.Println("inserting fight")
		i.db.CreateFight(fight)
	}
}
