package main

import (
	"os"
	"time"

	"github.com/geofpwhite/ufc_db_go/pkg/scraper"
)

func main() {
	s := scraper.NewScraper()
	s.Scrape()
	os.Rename("app.log", "logs/at"+time.Now().Local().String())
}
