package fetcher

import (
	"context"
	"fmt"
	"hackathon-api/internal/entity"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type MLHFetcher struct{}

func NewMLHFetcher() *MLHFetcher {
	return &MLHFetcher{}
}

func (f *MLHFetcher) Name() string {
	return "mlh"
}

func (f *MLHFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {

	year := time.Now().Year()
	if time.Now().Month() >= time.July {
		year++
	}
	url := fmt.Sprintf("https://mlh.io/seasons/%d/events", year)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", GetRandomUserAgent())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mlh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("mlh returned status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mlh html: %w", err)
	}

	var results []entity.Hackathon

	doc.Find(".event-wrapper").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".event-name").Text())
		if title == "" {
			return
		}

		link, _ := s.Find("a.event-link").Attr("href")
		location := strings.TrimSpace(s.Find(".event-location").Text())
		dateStr := strings.TrimSpace(s.Find(".event-date").Text())

		sDate, eDate := f.parseDate(dateStr, year)

		h := entity.Hackathon{
			ID:        fmt.Sprintf("mlh-%s", title),
			Title:     title,
			Platform:  "mlh",
			URL:       link,
			Location:  location,
			Country:   "India",
			StartDate: sDate,
			EndDate:   eDate,
			Mode:      "offline",
		}

		if strings.Contains(strings.ToLower(location), "global") || strings.Contains(strings.ToLower(title), "global") {
			h.Mode = "online"
		}

		h.Country = location

		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	})

	return results, nil
}

func (f *MLHFetcher) parseDate(dateStr string, seasonYear int) (time.Time, time.Time) {

	cleaner := strings.NewReplacer("st", "", "nd", "", "rd", "", "th", "")
	clean := cleaner.Replace(dateStr)

	parts := strings.Split(clean, "-")
	if len(parts) == 0 {
		return time.Time{}, time.Time{}
	}

	startPart := strings.TrimSpace(parts[0])

	t, err := time.Parse("Jan 2", startPart)
	if err != nil {
		return time.Time{}, time.Time{}
	}

	y := seasonYear
	if t.Month() >= time.August {
		y = seasonYear - 1
	}
	sDate := time.Date(y, t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	var eDate time.Time
	if len(parts) > 1 {
		endPart := strings.TrimSpace(parts[1])

		if len(endPart) <= 2 {
			day, _ := time.Parse("2", endPart)
			if !day.IsZero() {
				eDate = time.Date(y, t.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
			}
		} else {

			t2, err := time.Parse("Jan 2", endPart)
			if err == nil {
				y2 := seasonYear
				if t2.Month() >= time.August {
					y2 = seasonYear - 1
				}
				eDate = time.Date(y2, t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC)
			}
		}
	} else {
		eDate = sDate
	}

	return sDate, eDate
}
