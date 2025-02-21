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
	go func() {
		c := colly.NewCollector()
		c2 := colly.NewCollector()
		db := database.Init()
		ctx := colly.NewContext()
		visitedEvents, visitedFights := make(map[string]bool), make(map[string]bool)
		fmt.Println(visitedFights)
		c2.OnHTML("i.b-fight-details__text-item_first", func(e *colly.HTMLElement) {
			fields := strings.Fields(e.Text)
			if fields[0] == "Method:" {
				ctx.Put("method", fields[1])
			}
		})
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
			wIndex, lIndex, dIndex, ncIndex := slices.Index(fields, "W"), slices.Index(fields, "L"), slices.Index(fields[1:], "D"), slices.Index(fields[1:], "NC")
			var wInfo, lInfo []string
			if dIndex != -1 {
				//fight was a draw
				wInfo, lInfo = fields[:dIndex+1], fields[dIndex+1:]
				fs1.Outcome, fs2.Outcome = "DRAW", "DRAW"
			} else if ncIndex != -1 {
				wInfo, lInfo = fields[:ncIndex+1], fields[ncIndex+1:]
				fs1.Outcome, fs2.Outcome = "NC", "NC"
			} else {
				fmt.Println(fields)
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
			f.Fighter1Stats, f.Fighter2Stats = fs1, fs2
			f.EventID = event.ID
			method := ctx.Get("method")
			fmt.Println(method, "method")
			switch method {
			case "KO/TKO", "TKO", "KO", "Submission":
				f.Fighter1Stats.Outcome = "WINF"
				f.Fighter2Stats.Outcome = "LOSEF"
			case "Decision":
				f.Fighter1Stats.Outcome = "WIND"
				f.Fighter2Stats.Outcome = "LOSED"
			}
			ctx.Put("round", "f")
			lastFight := ctx.GetAny("fight")
			if lastFight != nil {
				fmt.Println("lastfight", lastFight)
				fightChannel <- *(lastFight.(*model.Fight))
				fmt.Println("sent through channel", len(fightChannel))
			}
			ctx.Put("fight", f)
		})
		c2.OnHTML("thead", func(e *colly.HTMLElement) {
			fmt.Println(joinStrings(strings.Fields(e.Text)), "thead")
			fields := strings.Fields(e.Text)
			if fields[0] == "Round" {
				ctx.Put("max round", fields[1])
			}
		})
		c2.OnHTML("tbody", func(e *colly.HTMLElement) {
			f := ctx.GetAny("fight").(*model.Fight)
			r := ctx.Get("round")
			// fmt.Println("fighter", f)
			x := joinStrings(strings.Fields(e.Text))
			if len(x) < 4 {
				fmt.Println(x)
				return
			}
			index := slices.IndexFunc(x, func(s string) bool {
				_, err := strconv.Atoi(s)
				if err == nil {
					return true
				}
				return false
			})
			if index == -1 {
				return
			} else {
				if index > slices.IndexFunc(x, func(s string) bool {
					return strings.Contains(s, "of")
				}) {
					return
				}
			}
			x = x[index:]
			x = slices.DeleteFunc(x, func(s string) bool {
				return strings.Contains(s, "%") || strings.Contains(s, "---")
			})

			//[kd1 kd2 sigstr1 sigstr2 totalstr1 totalstr2 td1 td2 subatt1 subatt2 rev1 rev2 ctrl1 ctrl2 ]
			nums := stringSliceToIntSlice(x)
			finalRound, err := strconv.Atoi(ctx.Get("max round"))
			if err != nil {
				fmt.Println(err)
				return
			}
			f.Fighter1Stats.FinalRound = finalRound
			f.Fighter2Stats.FinalRound = finalRound
			var kd1, kd2, sigstr1a,
				sigstr2a,
				sigstr1l,
				sigstr2l,
				totalstr1a,
				totalstr2a,
				totalstr1l,
				totalstr2l,
				td1a,
				td2a,
				td1l,
				td2l,
				subatt1,
				subatt2,
				_,
				_,
				ctrl1,
				ctrl2 int = nums[0], nums[1], nums[2], nums[3], nums[4], nums[5], nums[6], nums[7], nums[8], nums[9], nums[10], nums[11], nums[12], nums[13], nums[14], nums[15], nums[16], nums[17], nums[18], nums[19]

			switch r {
			case "f":
				f.Fighter1Stats.Knockdowns = kd1
				f.Fighter2Stats.Knockdowns = kd2
				f.Fighter1Stats.SigStrikesThrown = sigstr1a
				f.Fighter2Stats.SigStrikesThrown = sigstr2a
				f.Fighter1Stats.SigStrikesLanded = sigstr1l
				f.Fighter2Stats.SigStrikesLanded = sigstr2l
				f.Fighter1Stats.TotalStrikesThrown = totalstr1a
				f.Fighter2Stats.TotalStrikesThrown = totalstr2a
				f.Fighter1Stats.TotalStrikesLanded = totalstr1l
				f.Fighter2Stats.TotalStrikesLanded = totalstr2l
				f.Fighter1Stats.TakedownsAttempted = td1a
				f.Fighter2Stats.TakedownsAttempted = td2a
				f.Fighter1Stats.TakedownsLanded = td1l
				f.Fighter2Stats.TakedownsLanded = td2l
				f.Fighter1Stats.SubAttempts = subatt1
				f.Fighter2Stats.SubAttempts = subatt2
				f.Fighter1Stats.ControlTime = ctrl1
				f.Fighter2Stats.ControlTime = ctrl2
				ctx.Put("round", "1")
			default:
				num, err := strconv.Atoi(r)
				if err != nil {
					fmt.Println(err)
					return
				}
				rs1 := model.RoundStats{
					Round:              num,
					FightStatsID:       f.Fighter1Stats.ID,
					Knockdowns:         kd1,
					SigStrikesThrown:   sigstr1a,
					SigStrikesLanded:   sigstr1l,
					TotalStrikesThrown: totalstr1a,
					TotalStrikesLanded: totalstr1l,
					TakedownsAttempted: td1a,
					TakedownsLanded:    td1l,
					SubAttempts:        subatt1,
					ControlTime:        ctrl1,
				}
				rs2 := model.RoundStats{
					Round:              num,
					FightStatsID:       f.Fighter2Stats.ID,
					Knockdowns:         kd2,
					SigStrikesThrown:   sigstr2a,
					SigStrikesLanded:   sigstr2l,
					TotalStrikesThrown: totalstr2a,
					TotalStrikesLanded: totalstr2l,
					TakedownsAttempted: td2a,
					TakedownsLanded:    td2l,
					SubAttempts:        subatt2,
					ControlTime:        ctrl2,
				}
				f.Fighter1Stats.RoundStats = append(f.Fighter1Stats.RoundStats, rs1)
				f.Fighter2Stats.RoundStats = append(f.Fighter2Stats.RoundStats, rs2)
				ctx.Put("round", strconv.Itoa(num+1))
			}
			fmt.Println(stringSliceToIntSlice(x))
			// fmt.Println(kd1, kd2, err, err2)
			// if err != nil || err2 != nil {
			// 	return
			// }
			ctx.Put("fight", f)
		})
		// c2.OnHTML("div.b-fight-details__fight", func(e *colly.HTMLElement) {
		// 	fmt.Println(strings.Fields(e.Text))
		// })

		c.Visit("http://ufcstats.com/statistics/events/completed?page=all")
	}()
	return fightChannel
}

func stringSliceToIntSlice(s []string) []int {
	ret := make([]int, 0, len(s))
	for _, str := range s {
		if strings.Contains(str, ":") {
			strs := strings.Split(str, ":")
			minutes, err := strconv.Atoi(strs[0])
			if err != nil {
				continue
			}
			seconds, err := strconv.Atoi(strs[1])
			if err != nil {
				continue
			}
			ret = append(ret, (minutes*60)+seconds)
			continue
		}
		strs := strings.Split(str, "of")
		for _, str2 := range strs {
			num, err := strconv.Atoi(str2)
			if err != nil {
				continue
			}
			ret = append(ret, num)
		}
	}
	return ret
}

func joinStrings(s []string) []string {
	var ret []string = make([]string, 0, len(s))
	for i := 0; i < len(s); i++ {
		str := s[i]
		if str == "of" {
			ret[len(ret)-1] = ret[len(ret)-1] + str + s[i+1]
			i++
		} else {
			ret = append(ret, str)
		}
	}
	return ret
}
