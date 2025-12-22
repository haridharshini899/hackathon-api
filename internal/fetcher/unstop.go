package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"hackathon-api/internal/entity"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type UnstopFetcher struct{}

func NewUnstopFetcher() *UnstopFetcher {
	return &UnstopFetcher{}
}

func (f *UnstopFetcher) Name() string {
	return "unstop"
}

func (f *UnstopFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {
	url := "https://unstop.com/api/public/opportunity/search-result?opportunity=hackathons&page=1&per_page=200&oppstatus=open"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", GetRandomUserAgent())
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unstop: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unstop returned status: %d", resp.StatusCode)
	}

	type unstopHackathon struct {
		ID         int    `json:"id"`
		Title      string `json:"title"`
		SEOUrl     string `json:"seo_url"`
		StartDates string `json:"start_date"`
		EndDates   string `json:"end_date"`
		Region     string `json:"region"`
		Location   string `json:"location"`
	}

	type unstopResponse struct {
		Data struct {
			Data []unstopHackathon `json:"data"`
		} `json:"data"`
	}

	var apiResp unstopResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode unstop json: %w", err)
	}

	var results []entity.Hackathon
	for _, raw := range apiResp.Data.Data {

		mode := "offline"
		if strings.Contains(strings.ToLower(raw.Location), "online") {
			mode = "online"
		}

		parseDate := func(s string) time.Time {
			if s == "" {
				return time.Time{}
			}
			layouts := []string{
				"2006-01-02 15:04:05",
				time.RFC3339,
				"2006-01-02T15:04:05",
			}
			for _, l := range layouts {
				if t, err := time.Parse(l, s); err == nil {
					return t
				}
			}
			return time.Time{}
		}

		start := parseDate(raw.StartDates)
		end := parseDate(raw.EndDates)

		h := entity.Hackathon{
			ID:        fmt.Sprintf("unstop-%d", raw.ID),
			Title:     raw.Title,
			Platform:  "unstop",
			URL:       raw.SEOUrl,
			Country:   "India",
			Mode:      mode,
			Location:  raw.Location,
			StartDate: start,
			EndDate:   end,
		}

		if len(h.URL) > 0 && h.URL[0] == '/' {
			h.URL = "https://unstop.com" + h.URL
		}

		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for i := range results {
		h := &results[i]
		if h.StartDate.IsZero() {
			wg.Add(1)
			sem <- struct{}{}
			go func(hackathon *entity.Hackathon) {
				defer wg.Done()
				defer func() { <-sem }()
				f.enrichDetails(ctx, hackathon)
			}(h)
		} else {
		}
	}
	wg.Wait()

	return results, nil
}

func (f *UnstopFetcher) enrichDetails(ctx context.Context, h *entity.Hackathon) {

	parts := strings.Split(h.ID, "-")
	if len(parts) < 2 {
		return
	}
	numericID := parts[1]

	url := fmt.Sprintf("https://unstop.com/api/public/opportunity/search-result?id=%s", numericID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return
	}

	var detail map[string]interface{}

	if d1, ok := raw["data"].(map[string]interface{}); ok {
		if d2, ok := d1["data"].([]interface{}); ok && len(d2) > 0 {
			if item, ok := d2[0].(map[string]interface{}); ok {
				detail = item
			}
		}
	}

	if detail == nil {
		return
	}

	startDateStr, _ := detail["start_date"].(string)
	endDateStr, _ := detail["end_date"].(string)

	if startDateStr == "" {
		s := string(bodyBytes)

		keys := []string{`"start_date":"`, `"startDate":"`, `"start_regn_dt":"`}
		for _, key := range keys {
			if idx := strings.Index(s, key); idx != -1 {
				valStart := idx + len(key)
				valEnd := strings.Index(s[valStart:], `"`)
				if valEnd != -1 {
					startDateStr = s[valStart : valStart+valEnd]
					break
				}
			}
		}
	}

	parseUnstopDate := func(s string) time.Time {
		if s == "" {
			return time.Time{}
		}
		sISO := strings.Replace(s, " ", "T", 1)
		layouts := []string{
			"2006-01-02 15:04:05-07:00",
			time.RFC3339,
			"2006-01-02T15:04:05-07:00",
			"2006-01-02 15:04:05",
		}
		for _, l := range layouts {
			if t, err := time.Parse(l, s); err == nil {
				return t
			}
			if t, err := time.Parse(l, sISO); err == nil {
				return t
			}
		}
		return time.Time{}
	}

	if t := parseUnstopDate(startDateStr); !t.IsZero() {
		h.StartDate = t
	}
	if t := parseUnstopDate(endDateStr); !t.IsZero() {
		h.EndDate = t
	}
}
