package golfers

import (
	"errors"
	"time"
)

var (
	ErrInvalidGolferID = errors.New("invalid golfer ID")
	ErrGolferNotFound  = errors.New("golfer not found")
)

type GolferDetail struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	FirstName         string     `json:"first_name,omitempty"`
	LastName          string     `json:"last_name,omitempty"`
	CountryCode       string     `json:"country_code"`
	Country           string     `json:"country,omitempty"`
	OWGR              *int       `json:"owgr,omitempty"`
	ImageURL          string     `json:"image_url,omitempty"`
	Active            bool       `json:"active"`
	Height            string     `json:"height,omitempty"`
	Weight            string     `json:"weight,omitempty"`
	BirthDate         *time.Time `json:"birth_date,omitempty"`
	BirthplaceCity    string     `json:"birthplace_city,omitempty"`
	BirthplaceState   string     `json:"birthplace_state,omitempty"`
	BirthplaceCountry string     `json:"birthplace_country,omitempty"`
	TurnedPro         *int       `json:"turned_pro,omitempty"`
	School            string     `json:"school,omitempty"`
	ResidenceCity     string     `json:"residence_city,omitempty"`
	ResidenceState    string     `json:"residence_state,omitempty"`
	ResidenceCountry  string     `json:"residence_country,omitempty"`
}

type GetGolferResponse struct {
	Golfer GolferDetail `json:"golfer"`
}
