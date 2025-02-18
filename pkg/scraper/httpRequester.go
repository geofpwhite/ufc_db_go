/*
 */
package scraper

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/geofpwhite/ufc_db_go/pkg/database"
	"github.com/geofpwhite/ufc_db_go/pkg/model"
	"github.com/gocolly/colly"
)

const CMCOEFF = 2.54
const INCOEFF = 0.39370079
const KGCOEFF = 0.45359237
const LBSCOEFF = 2.20462262

func SendEventResponses() <-chan model.Event {

	eventChannel := make(chan model.Event)
	cur := model.Event{}
	go func() {
		c := colly.NewCollector() // Find and visit all links
		c.OnHTML("td[class^='b-statistics__table-col']", func(e *colly.HTMLElement) {
			if e.Text == "" {
				return
			}
			fString := strings.Replace(e.Text, "\n", "", -1)
			fString = strings.Replace(fString, "            ", "|", -1)
			fString = strings.ReplaceAll(fString, "||||||  ", "|")
			fString = strings.Trim(fString, "| \r\n\t")

			if barIndex := strings.Index(fString, "|"); barIndex != -1 {
				cur.Title = fString[:barIndex]
				dateStr := strings.Trim(fString[barIndex+1:], " \n\r\t")
				date, err := time.Parse("January 2, 2006", dateStr)
				if err != nil {
					//handle this somehow
					fmt.Println(err)
				}
				cur.Date = date
				return
			}
			// fmt.Printf("%x", fString)
			cur.Location = strings.Trim(e.Text, " \r\n\t")
			eventChannel <- cur
		})
		c.Visit("http://ufcstats.com/statistics/events/completed?page=all")
		close(eventChannel)
	}()
	return eventChannel
}

func SendFighterResponses() <-chan model.Fighter {
	fighterChannel := make(chan model.Fighter)
	fighterAndDateChannel := make(chan model.Fighter)
	c, c2 := colly.NewCollector(), colly.NewCollector()
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	visited := make(map[string]bool)
	c.OnHTML("tr.b-statistics__table-row", func(e *colly.HTMLElement) {
		time.Sleep(1 * time.Millisecond)
		var (
			cur           *model.Fighter = new(model.Fighter)
			height, reach *int
			dob           time.Time
			first, last   string
			nickname      *string
		)
		fString := strings.ReplaceAll(e.Text, "\n", "|")
		fString = strings.ReplaceAll(fString, "  ", "")
		info := strings.Split(strings.Trim(fString, "|"), "|||")
		if len(info) == 9 { //this only happens if they have one name rather than a first and last, so we shall treat their one name as their last and their first will be blank.
			info = append(make([]string, 1, 10), info...)
		}
		if len(info) < 3 || info[0] == "First" {
			return
		}
		defer func() {
			go func() {
				*cur = model.Fighter{
					FirstName: first,
					LastName:  last,
					NickName:  nickname,
					DOB:       dob,
					Height:    height,
					Reach:     reach,
				}
				fighterChannel <- *cur
			}()
		}()
		first, last, nickname = info[0], info[1], &info[2]
		if *nickname == "--" || *nickname == "" {
			nickname = nil
		}
		feetAndInches := strings.Split(info[3], " ")
		if feetAndInches[0] == "--" {
			height = nil
		} else {
			feet, err := strconv.Atoi(strings.Trim(feetAndInches[0], " '\""))
			if err != nil {
				fmt.Println(err, feetAndInches)
			}
			inches, err := strconv.Atoi(strings.Trim(feetAndInches[1], " '\""))
			if err != nil || feet == 0 {
				fmt.Println(err, feetAndInches)
				feet, inches = 0, 0
			}
			inches += feet * 12
			cm := (int(float64(inches) * CMCOEFF))
			height = &cm
			if cm == 0 {
				height = nil
			}
		}
		switch info[5][:2] {
		case "--":
			reach = nil
		default:
			inches, err := strconv.Atoi(info[5][:2])
			if err != nil {
				reach = nil
				break
			}
			cm := int(float64(inches) * CMCOEFF)
			reach = &cm
		}
	})

	c.OnHTML("a[href^='http://ufcstats.com/fighter-details']", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		fmt.Println(e.Text)
		if !visited[url] {
			visited[url] = true
			c2.Visit(url)
		}
	})
	c2.OnHTML("li.b-list__box-list-item.b-list__box-list-item_type_block", func(e *colly.HTMLElement) {
		if e.Text == "" {
			return
		}
		if index := strings.Index(e.Text, "DOB:"); index != -1 {
			dobText := strings.TrimSpace(e.Text[index+4:])
			fighter := <-fighterChannel
			date, err := time.Parse("Jan 2, 2006", dobText)
			if err != nil {
				fmt.Println(err)
			}
			fighter.DOB = date
			fighterAndDateChannel <- fighter
		}
	})
	go func() {
		for _, char := range alphabet {
			c.Visit(fmt.Sprintf("http://ufcstats.com/statistics/fighters?char=%s&page=all", string(char)))
		}
		println("closed")
		close(fighterAndDateChannel)
	}()
	return fighterAndDateChannel
}

func SendFightResponses() <-chan model.Fight {
	fightChannel := make(chan model.Fight)
	c := colly.NewCollector()
	c2 := colly.NewCollector()
	db := database.Init()
	ctx := colly.NewContext()
	visitedEvents, visitedFights := make(map[string]bool), make(map[string]bool)
	fmt.Println(visitedFights)
	c.OnHTML("a[href^='http://ufcstats.com/event-details']", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		text := strings.Trim(e.Text, " \n\r\t")
		event, err := db.GetEventByTitle(text)
		fmt.Println(event, err)
		if err != nil {
			panic(err)
		}
		if !visitedEvents[url] {
			visitedEvents[url] = true
			c.Visit(url)
		}
		ctx.Put("event", event)
		fmt.Println("event", event)
	})
	c.OnHTML("a[href^='http://ufcstats.com/fight-details']", func(e *colly.HTMLElement) {
		event := ctx.GetAny("event").(*model.Event)
		fmt.Println(event.Title)
		c2.Visit(e.Attr("href"))
	})

	c2.OnHTML("div.b-fight-details__persons.clearfix", func(e *colly.HTMLElement) {
		event := ctx.GetAny("event").(*model.Event)
		fields := strings.Fields(e.Text)
		fs1 := model.FightStats{}
		fs2 := model.FightStats{}
		f := &model.Fight{EventID: event.ID, Fighter1Stats: fs1, Fighter2Stats: fs2}
		wIndex, lIndex, dIndex := slices.Index(fields, "W"), slices.Index(fields, "L"), slices.Index(fields[1:], "D")
		var wInfo, lInfo []string
		if dIndex != -1 {
			//fight was a draw
			wInfo, lInfo = fields[:dIndex+1], fields[dIndex+1:]
		} else {
			wInfo, lInfo = fields[:max(wIndex, lIndex)], fields[max(wIndex, lIndex):]
		}
		if wInfo[0] == "L" {
			wInfo, lInfo = lInfo, wInfo
		}
		wnicknameIndex := slices.IndexFunc(wInfo, func(s string) bool {
			return strings.Contains(s, "\"")
		})
		var wname, lname []string
		if wnicknameIndex == -1 {
			wname = wInfo[1:]
		} else {
			wname = wInfo[1:wnicknameIndex]
		}
		lnicknameIndex := slices.IndexFunc(lInfo, func(s string) bool {
			return strings.Contains(s, "\"")
		})
		if lnicknameIndex == -1 {
			lname = lInfo[1:]
		} else {
			lname = lInfo[1:lnicknameIndex]
		}
		var f1, f2 *model.Fighter
		var err error
		switch len(wname) {
		case 2:
			f1, err = db.GetFighterByFirstAndLastName(wname[0], wname[1])
		case 1:
			f1, err = db.GetFighterByFirstAndLastName("", wname[0])
		case 0:
			fmt.Println("no name")
			return
		default:
			lastName := strings.Join(wname[1:], " ")
			f1, err = db.GetFighterByFirstAndLastName(wname[0], lastName)
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		switch len(lname) {
		case 2:
			f2, err = db.GetFighterByFirstAndLastName(lname[0], lname[1])
		case 1:
			f2, err = db.GetFighterByFirstAndLastName("", lname[0])
		case 0:
			fmt.Println("no name")
			return
		default:
			lastName := strings.Join(lname[1:], " ")
			f2, err = db.GetFighterByFirstAndLastName(lname[0], lastName)
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(wInfo, lInfo)
		fmt.Println(f, fs1, fs2, f1, f2)
		fs1.FighterID, fs2.FighterID = f1.ID, f2.ID

	})
	c2.OnHTML("section.b-fight-details__section.js-fight-section", func(e *colly.HTMLElement) {
		fmt.Println(strings.Fields(e.Text))
	})

	c.Visit("http://ufcstats.com/statistics/events/completed?page=all")

	return fightChannel
}
