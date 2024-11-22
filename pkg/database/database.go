package database

import (
	"errors"
	"sync"
	"time"

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

func (database *Database) GetFighterByName(name string) (*model.Fighter, error) {
	database.mut.Lock()
	defer database.mut.Unlock()
	fighter := model.Fighter{Name: name}
	var result []model.Fighter
	ret := database.db.Where(&fighter).Find(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return nil, ret.Error
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

func (database *Database) GetAllFightsBetween(f1, f2 model.FightStats) ([]model.Fight, error) {
	database.mut.Lock()
	defer database.mut.Unlock()
	var result []model.Fight
	ret := database.db.Where(&model.Fight{Fighter1Stats: f1, Fighter2Stats: f2}).Find(&result)

	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return nil, ret.Error
	}
	return result, nil
}

func (database *Database) CreateFighter(name string, dob time.Time) error {
	database.mut.Lock()
	defer database.mut.Unlock()
	f := model.Fighter{Name: name, DOB: dob}
	var result model.Fighter
	ret := database.db.Where(&f).First(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return ret.Error
	}
	database.db.Save(&f)
	return nil
}

func (database *Database) CreateFight(f1, f2 model.FightStats, event model.Event) error {
	database.mut.Lock()
	defer database.mut.Unlock()
	f := model.Fight{Fighter1Stats: f1, Fighter2Stats: f2, EventID: event.ID}
	var result model.Fight
	ret := database.db.Where(&f).First(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return ret.Error
	}
	database.db.Save(&f)
	return nil
}

func (database *Database) CreateEvent(title string, date time.Time) error {
	database.mut.Lock()
	defer database.mut.Unlock()
	e := model.Event{Title: title, Date: date}
	var result model.Fighter
	ret := database.db.Where(&e).Find(&result)
	if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
		return ret.Error
	}
	database.db.Save(&e)
	return nil
}
