package database

import (
	"errors"
	"fmt"
	"sync"

	"github.com/geofpwhite/ufc_db_go/pkg/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	db  *gorm.DB
	mut *sync.Mutex
}

func Init() *Database {
	db, err := gorm.Open(sqlite.Open("ufc.db"))
	if err != nil {
		panic("error opening ufc.db")
	}
	database := Database{
		db:  db,
		mut: &sync.Mutex{},
	}
	database.db.AutoMigrate(new(model.FightStats), new(model.Fighter), new(model.Event), &model.Fight{})
	return &database
}

func (database *Database) GetFighterByFirstAndLastName(firstName, lastName string) (*model.Fighter, error) {
	database.mut.Lock()
	defer database.mut.Unlock()
	fighter := model.Fighter{FirstName: firstName, LastName: lastName}
	var result []model.Fighter
	ret := database.db.Where(&fighter).Find(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return nil, ret.Error
	}
	if len(result) == 0 {
		return nil, errors.New(fmt.Sprintf("no fighter found for %s %s", firstName, lastName))
	}
	return &result[0], nil
}
func (database *Database) GetFighterByID(id uint) (*model.Fighter, error) {
	database.mut.Lock()
	defer database.mut.Unlock()
	fighter := model.Fighter{ID: id}
	var result model.Fighter
	ret := database.db.Where(&fighter).First(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return nil, ret.Error
	}
	return &result, nil
}

func (database *Database) GetAllFightsBetween(f1, f2 model.Fighter) ([]model.Fight, error) {
	database.mut.Lock()
	defer database.mut.Unlock()
	var result []model.Fight
	ret := database.db.Where(&model.Fight{
		Fighter1Stats: model.FightStats{
			FighterID: f1.ID,
		},
		Fighter2Stats: model.FightStats{
			FighterID: f2.ID,
		},
	}).Find(&result)
	// ret := database.db.Exec(fmt.Sprintf("SELECT * FROM fights join fight_stats ON fights.id=fight_stats.fight_id where fighter_id=%d or fighter_id=%d", f1.ID, f2.ID))

	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return nil, ret.Error
	}
	return result, nil
}

func (database *Database) CreateFighter(f model.Fighter) error {
	database.mut.Lock()
	defer database.mut.Unlock()
	var result model.Fighter
	ret := database.db.Where(&f).First(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		database.db.Save(&f)
		return nil
	}
	return ret.Error
}

func (database *Database) CreateFight(fight model.Fight) error {
	database.mut.Lock()
	defer database.mut.Unlock()
	var result model.Fight
	ret := database.db.Where(&fight).First(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return ret.Error
	}
	database.db.Save(&fight)
	return nil
}

func (database *Database) CreateEvent(e model.Event) error {
	database.mut.Lock()
	defer database.mut.Unlock()
	var result model.Event
	ret := database.db.Where(&model.Event{Title: e.Title}).Find(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return ret.Error
	}
	database.db.Save(&e)
	return nil
}

func (database *Database) GetEventByTitle(title string) (*model.Event, error) {
	database.mut.Lock()
	defer database.mut.Unlock()
	var result model.Event
	ret := database.db.Where(&model.Event{Title: title}).Find(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return nil, ret.Error
	}
	return &result, nil
}
