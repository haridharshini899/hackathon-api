package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"hackathon-api/internal/entity"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Hack2SkillFetcher struct{}

func NewHack2SkillFetcher() *Hack2SkillFetcher {
	return &Hack2SkillFetcher{}
}

func (f *Hack2SkillFetcher) Name() string {
	return "hack2skill"
}

func (f *Hack2SkillFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {

	results, err := f.fetchAPI(ctx)
	if err == nil && len(results) > 0 {
		return results, nil
	}

	fmt.Println("Hack2Skill API returned 0 results, falling back to scraping.")
	return f.scrapeHTML(ctx)
}

func (f *Hack2SkillFetcher) fetchAPI(ctx context.Context) ([]entity.Hackathon, error) {

	start := time.Now().Format("2006-01-02T15:04:05.000Z")
	end := time.Now().AddDate(2, 0, 0).Format("2006-01-02T15:04:05.000Z")

	url := fmt.Sprintf("https://vision.hack2skill.com/api/v1/innovator/public/event/public-list?page=1&records=50&search=&start=%s&end=%s", start, end)

	req, err := CreateRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", "https://vision.hack2skill.com/hackathons-listing")
	req.Header.Set("Origin", "https://vision.hack2skill.com")

	client := NewClient()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hack2skill: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("hack2skill returned status: %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode hack2skill json: %w", err)
	}

	var docs []interface{}

	if data, ok := raw["data"].(map[string]interface{}); ok {
		if d, ok := data["docs"].([]interface{}); ok {
			docs = d
		}
	} else if d, ok := raw["data"].([]interface{}); ok {
		docs = d
	}

	var results []entity.Hackathon
	for _, item := range docs {
		rawMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := rawMap["title"].(string)
		slug, _ := rawMap["eventUrl"].(string)
		modeRaw, _ := rawMap["mode"].(string)
		city, _ := rawMap["city"].(string)
		country, _ := rawMap["country"].(string)

		var idStr string
		if val, ok := rawMap["id"]; ok && val != nil {
			idStr = fmt.Sprintf("%v", val)
		} else if val, ok := rawMap["_id"]; ok && val != nil {
			idStr = fmt.Sprintf("%v", val)
		}

		var sDate, eDate time.Time

		parseTime := func(v interface{}) time.Time {
			s, ok := v.(string)
			if !ok {
				return time.Time{}
			}

			formats := []string{time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05Z", "2006-01-02"}
			for _, f := range formats {
				if t, err := time.Parse(f, s); err == nil {
					return t
				}
			}
			return time.Time{}
		}

		if configs, ok := rawMap["configs"].(map[string]interface{}); ok {
			if s, ok := configs["slug"].(string); ok && slug == "" {
				slug = s
			}
		}

		regStart := parseTime(rawMap["registrationStart"])
		regEnd := parseTime(rawMap["registrationEnd"])
		subStart := parseTime(rawMap["submissionStart"])
		subEnd := parseTime(rawMap["submissionEnd"])

		if !subStart.IsZero() {
			sDate = subStart
		} else {
			sDate = regStart
		}

		if !subEnd.IsZero() {
			eDate = subEnd
		} else {
			eDate = regEnd
		}

		if slug == "" {
			slug, _ = rawMap["slug"].(string)
		}
		if slug == "" {
			slug, _ = rawMap["event_slug"].(string)
		}

		mode := "offline"
		if strings.EqualFold(modeRaw, "online") || strings.EqualFold(modeRaw, "virtual") {
			mode = "online"
		} else if strings.EqualFold(modeRaw, "hybrid") {
			mode = "hybrid"
		}

		eventUrl := "https://vision.hack2skill.com/event/"
		if slug != "" {
			eventUrl = fmt.Sprintf("https://vision.hack2skill.com/event/%s", slug)
		} else if idStr != "" {
			eventUrl = fmt.Sprintf("https://vision.hack2skill.com/event/%s", idStr)
		}

		h := entity.Hackathon{
			ID:              fmt.Sprintf("h2s-%s", idStr),
			Title:           title,
			Platform:        "hack2skill",
			Mode:            mode,
			Location:        city,
			Country:         "India",
			StartDate:       sDate,
			EndDate:         eDate,
			RegistrationEnd: regEnd,
			URL:             eventUrl,
		}

		if country != "" {
			h.Country = country
		}

		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	}

	return results, nil
}

func (f *Hack2SkillFetcher) scrapeHTML(ctx context.Context) ([]entity.Hackathon, error) {
	url := "https://hack2skill.com/hackathons"
	req, err := CreateRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := NewClient()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []entity.Hackathon

	doc.Find("div.new-card, div.event-card, a[href*='/hackathons/']").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}

		title := strings.TrimSpace(s.Find("h3, .title, .event-title").Text())
		if title == "" {
			title = strings.TrimSpace(s.Text())
		}
		if len(title) < 5 {
			return
		}

		h := entity.Hackathon{
			ID:        fmt.Sprintf("h2s-scrape-%d", i),
			Title:     title,
			Platform:  "hack2skill",
			Mode:      "online",
			Country:   "India",
			URL:       link,
			StartDate: time.Time{},
		}
		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	})

	return results, nil
}
