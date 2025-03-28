package fighter

import "time"

type Fighter struct {
	ID                          uint      `json:"id"`
	FirstName                   string    `json:"first_name"`
	LastName                    string    `json:"last_name"`
	NickName                    string    `json:"nick_name,omitempty"`
	DOB                         time.Time `json:"dob"`
	Height                      int       `json:"height,omitempty"`
	Reach                       int       `json:"reach,omitempty"`
	AvgSigStrikesLandedPerRound float64   `json:"avg_sig_strikes_landed_per_round,omitempty"`
	Wins                        int       `json:"wins,omitempty"`
	Losses                      int       `json:"losses,omitempty"`
	NoContests                  int       `json:"no_contests,omitempty"`
	Draws                       int       `json:"draws,omitempty"`
}
