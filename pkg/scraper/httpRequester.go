/*
 */
package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
			// fmt.Println(e.Text)
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
	c, c2 := colly.NewCollector(), colly.NewCollector()
	visitedEvents, visitedFights := make(map[string]bool), make(map[string]bool)
	c.OnHTML("a[href^='http://ufcstats.com/event-details']", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		if !visitedEvents[url] {
			visitedEvents[url] = true
			c.Visit(url)
		}
	})
	c.OnHTML("a[href^='http://ufcstats.com/fight-details']", func(e *colly.HTMLElement) {
		c2.Visit(e.Attr("href"))
	})
	c.Visit("http://ufcstats.com/statistics/events/completed?page=all")

	return fightChannel
}

// func (hr *HttpRequester) SendFightStatsResponses()
