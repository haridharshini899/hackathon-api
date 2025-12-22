package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"hackathon-api/pkg/entity"
	"strings"
	"time"
)

type DevpostFetcher struct{}

func NewDevpostFetcher() *DevpostFetcher {
	return &DevpostFetcher{}
}

func (f *DevpostFetcher) Name() string {
	return "devpost"
}

func (f *DevpostFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {
	url := "https://devpost.com/api/hackathons"

	req, err := CreateRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Ch-Ua", "\"Google Chrome\";v=\"143\", \"Chromium\";v=\"143\", \"Not A(Brand\";v=\"24\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Windows\"")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	req.Header.Set("Origin", "https://devpost.com")

	client := NewClient()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devpost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("devpost returned status: %d", resp.StatusCode)
	}

	type devpostHackathon struct {
		ID               int         `json:"id"`
		Title            string      `json:"title"`
		URL              string      `json:"url"`
		SubmissionPeriod string      `json:"submission_period_dates"`
		PrizeAmount      string      `json:"prize_amount"`
		Location         interface{} `json:"displayed_location"`
	}

	type devpostResponse struct {
		Hackathons []devpostHackathon `json:"hackathons"`
	}

	var apiResp devpostResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode devpost json: %w", err)
	}

	var results []entity.Hackathon
	for _, raw := range apiResp.Hackathons {
		locStr := "Online"
		if s, ok := raw.Location.(string); ok {
			locStr = s
		} else if m, ok := raw.Location.(map[string]interface{}); ok {
			if l, ok := m["location"].(string); ok {
				locStr = l
			}
		}

		mode := "offline"
		if strings.EqualFold(locStr, "online") || strings.Contains(strings.ToLower(locStr), "online") {
			mode = "online"
		}

		sDate, eDate := f.parseDates(raw.SubmissionPeriod)

		idStr := fmt.Sprintf("devpost-%d", raw.ID)

		h := entity.Hackathon{
			ID:        idStr,
			Title:     raw.Title,
			Platform:  "devpost",
			Mode:      mode,
			Location:  locStr,
			Country:   "India",
			URL:       raw.URL,
			StartDate: sDate,
			EndDate:   eDate,
		}

		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	}

	return results, nil
}

func (f *DevpostFetcher) parseDates(period string) (time.Time, time.Time) {

	parts := strings.Split(period, "-")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	parseFull := func(s string) time.Time {
		t, err := time.Parse("Jan 02, 2006", s)
		if err == nil {
			return t
		}
		return time.Time{}
	}

	eDate := parseFull(endStr)
	if eDate.IsZero() {
		return time.Time{}, time.Time{}
	}

	sDate := parseFull(startStr)
	if !sDate.IsZero() {
		return sDate, eDate
	}

	t, err := time.Parse("Jan 02", startStr)
	if err == nil {

		sDate = time.Date(eDate.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

		if sDate.After(eDate) {
			sDate = sDate.AddDate(-1, 0, 0)
		}
		return sDate, eDate
	}

	return time.Time{}, eDate
}
