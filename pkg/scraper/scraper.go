package scraper

type Scraper struct {
	inserter Inserter
}

func NewScraper() *Scraper {
	return &Scraper{
		inserter: *NewInserter(),
	}
}

func (s *Scraper) Scrape() {
	// s.inserter.InsertEvents(SendEventResponses())
	// s.inserter.InsertFighters(SendFighterResponses())
	s.inserter.InsertFights(SendFightResponses())
	// s.inserter.InsertFightStats(SendFightStatResponses())
}
