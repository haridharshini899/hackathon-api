package entity

import (
	"encoding/json"
	"strings"
	"time"
)

type Hackathon struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Platform        string    `json:"platform"`
	Mode            string    `json:"mode"`
	Location        string    `json:"location"`
	Country         string    `json:"country"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	RegistrationEnd time.Time `json:"registration_end"`
	URL             string    `json:"url"`
	PrizePool       string    `json:"prize_pool,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
}

func (h *Hackathon) IsIndiaSpecific() bool {

	if strings.EqualFold(h.Country, "India") {
		return true
	}

	indianCities := []string{
		"bangalore", "bengaluru", "delhi", "mumbai", "pune", "hyderabad",
		"chennai", "gurgaon", "gurugram", "noida", "kolkata", "ahmedabad",
		"jaipur", "indore", "chandigarh", "coimbatore", "kochi", "trivandrum",
	}
	loc := strings.ToLower(h.Location)
	for _, city := range indianCities {
		if strings.Contains(loc, city) {
			return true
		}
	}

	if strings.EqualFold(h.Mode, "online") {
		return true
	}

	return false
}

func (h *Hackathon) HasExpired() bool {
	if h.EndDate.IsZero() {
		return false
	}

	return h.EndDate.Add(24 * time.Hour).Before(time.Now())
}

func (h *Hackathon) MarshalJSON() ([]byte, error) {
	type Alias Hackathon

	aux := &struct {
		StartDate       interface{} `json:"start_date"`
		EndDate         interface{} `json:"end_date"`
		RegistrationEnd interface{} `json:"registration_end"`
		Location        string      `json:"location"`
		*Alias
	}{
		Alias: (*Alias)(h),
	}

	if h.StartDate.IsZero() {
		aux.StartDate = nil
	} else {
		aux.StartDate = h.StartDate.Format("2006-01-02")
	}

	if h.EndDate.IsZero() {
		aux.EndDate = nil
	} else {
		aux.EndDate = h.EndDate.Format("2006-01-02")
	}

	if h.RegistrationEnd.IsZero() {
		aux.RegistrationEnd = nil
	} else {
		aux.RegistrationEnd = h.RegistrationEnd.Format("2006-01-02")
	}

	if h.Location == "" {
		if strings.EqualFold(h.Mode, "online") {
			aux.Location = "Online"
		} else {
			aux.Location = "TBA"
		}
	} else {
		aux.Location = h.Location
	}

	return json.Marshal(aux)
}
