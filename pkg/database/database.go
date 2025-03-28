package database

import (
	_ "embed"
	"errors"
	"fmt"
	"sync"

	"github.com/geofpwhite/ufc_db_go/pkg/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed queries/avgSigStrikesPerRound.sql
var avgSigStrikesLandedPerRoundQuery string

//go:embed queries/allFightersWithWLD.sql
var allFightersQuery string

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
	database.db.AutoMigrate(new(model.FightStats), new(model.Fighter), new(model.Event), new(model.Fight), new(model.RoundStats))
	return &database
}

type fighterStats struct {
	ID        uint   `json:"id";db:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Height    *int   `json:"height"`
	Reach     *int   `json:"reach"`

	WinCount  int `json:"wins"`
	LossCount int `json:"losses"`
	DrawCount int `json:"draws"`
}

func (database *Database) GetAllFighters(page int) ([]fighterStats, error) {
	database.mut.Lock()
	defer database.mut.Unlock()
	var result []fighterStats = make([]fighterStats, 100)

	// res, err := sql.Query(allFightersQuery)
	fmt.Println(allFightersQuery)
	res, err := database.db.Raw(allFightersQuery, page*100).Rows()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// res, err := sql.Query("select * from fighters;")
	fmt.Println(res.Next())
	for i := 0; i < len(result) && res.Next(); i++ {

		err = res.Scan(&result[i].ID, &result[i].FirstName, &result[i].LastName, &result[i].Height, &result[i].Reach, &result[i].WinCount, &result[i].LossCount, &result[i].DrawCount)
		// err = res.Scan(&result[i].ID, &result[i].FirstName, &result[i].LastName, &result[i].Height, &result[i].Reach)
		fmt.Println(result[i])
		// err = res.Scan(&result[i])
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	return result, nil
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
		return nil, fmt.Errorf("no fighter found for %s %s", firstName, lastName)
	}
	return &result[0], nil
}

func (database *Database) GetAverageSigStrikesLandedPerRound(fighterID uint) (float64, error) {
	database.mut.Lock()
	defer database.mut.Unlock()
	var result float64
	sql, _ := database.db.DB()
	ret := sql.QueryRow(avgSigStrikesLandedPerRoundQuery, fighterID).Scan(&result)
	if ret != nil {
		return 0, ret
	}
	return result, nil
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
	database.db.Save(&f)
	return nil
}

func (database *Database) CreateFight(fight model.Fight) error {
	database.mut.Lock()
	defer database.mut.Unlock()
	database.db.Save(&fight)
	fmt.Println("saved fight")
	return nil
}

func (database *Database) CreateEvent(e model.Event) error {
	database.mut.Lock()
	defer database.mut.Unlock()
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
