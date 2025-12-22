package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"hackathon-api/internal/entity"
	"strings"
	"time"
)

type HackerEarthFetcher struct{}

func NewHackerEarthFetcher() *HackerEarthFetcher {
	return &HackerEarthFetcher{}
}

func (f *HackerEarthFetcher) Name() string {
	return "hackerearth"
}

func (f *HackerEarthFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {
	return f.fetchAPI(ctx)
}

func (f *HackerEarthFetcher) fetchAPI(ctx context.Context) ([]entity.Hackathon, error) {
	url := "https://www.hackerearth.com/api/community/challenges/compete/?limit=100&status=UPCOMING"

	req, err := CreateRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://www.hackerearth.com/challenges/hackathon/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := NewClient()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, formatError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("hackerearth api status: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Title    string `json:"title"`
			Slug     string `json:"slug"`
			Type     string `json:"type"`
			Start    string `json:"start"`
			End      string `json:"end"`
			URL      string `json:"url"`
			Location string `json:"location"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode hackerearth json: %w", err)
	}

	var results []entity.Hackathon
	for _, item := range result.Data {

		if !strings.EqualFold(item.Type, "Hackathon") {
			continue
		}

		sDate, _ := time.Parse("2006-01-02T15:04:05", item.Start)
		eDate, _ := time.Parse("2006-01-02T15:04:05", item.End)

		if eDate.Before(time.Now()) {
			continue
		}

		fullURL := item.URL
		if !strings.HasPrefix(fullURL, "http") {
			if item.Slug != "" {
				fullURL = fmt.Sprintf("https://www.hackerearth.com/challenges/hackathon/%s/", item.Slug)
			} else {
				fullURL = "https://www.hackerearth.com" + fullURL
			}
		}

		h := entity.Hackathon{
			ID:        fmt.Sprintf("he-%s", item.Slug),
			Title:     item.Title,
			Platform:  "hackerearth",
			Mode:      "online",
			Country:   "India",
			StartDate: sDate,
			EndDate:   eDate,
			URL:       fullURL,
		}

		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	}

	return results, nil
}

func formatError(err error) error {
	return fmt.Errorf("hackerearth error: %w", err)
}
