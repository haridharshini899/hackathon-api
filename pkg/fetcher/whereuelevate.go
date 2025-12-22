package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"hackathon-api/pkg/entity"
	"strings"
	"time"
)

type WhereUElevateFetcher struct{}

func NewWhereUElevateFetcher() *WhereUElevateFetcher {
	return &WhereUElevateFetcher{}
}

func (f *WhereUElevateFetcher) Name() string {
	return "whereuelevate"
}

func (f *WhereUElevateFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {
	url := "https://api.whereuelevate.com/internity/api/v1/drills/search?drillCategory=HACKATHON&drillId=all&limit=30&mode=all&offset=0&order=DESC&status=all&type=all&isActive=true&hideFromUserListing=false"

	req, err := CreateRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://whereuelevate.com/drills")
	req.Header.Set("Origin", "https://whereuelevate.com")

	client := NewClient()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch whereuelevate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("whereuelevate returned status: %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode whereuelevate json: %w", err)
	}

	var drills []interface{}

	if data, ok := raw["data"].(map[string]interface{}); ok {
		if d, ok := data["drills"].([]interface{}); ok {
			drills = d
		} else if d, ok := data["content"].([]interface{}); ok {
			drills = d
		}
	} else if d, ok := raw["data"].([]interface{}); ok {
		drills = d
	}

	var results []entity.Hackathon
	for _, item := range drills {
		rawMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := rawMap["title"].(string)
		if title == "" {
			title, _ = rawMap["drillTitle"].(string)
		}
		if title == "" {
			title, _ = rawMap["drillName"].(string)
		}

		slug, _ := rawMap["slug"].(string)
		if slug == "" {
			slug, _ = rawMap["drillCustUrl"].(string)
		}

		modeRaw, _ := rawMap["mode"].(string)
		if modeRaw == "" {
			modeRaw, _ = rawMap["drillMode"].(string)
		}

		location, _ := rawMap["location"].(string)

		startDate, _ := rawMap["startDate"].(string)
		if startDate == "" {
			startDate, _ = rawMap["drillStartDate"].(string)
		}
		if startDate == "" {
			startDate, _ = rawMap["drillStartDt"].(string)
		}

		endDate, _ := rawMap["endDate"].(string)
		if endDate == "" {
			endDate, _ = rawMap["drillEndDate"].(string)
		}
		if endDate == "" {
			endDate, _ = rawMap["drillEndDt"].(string)
		}

		regEndDate, _ := rawMap["drillRegistrationEndDt"].(string)

		idVal, _ := rawMap["drillId"].(float64)

		mode := "offline"
		if strings.EqualFold(modeRaw, "online") || strings.EqualFold(location, "online") {
			mode = "online"
		}

		parseUE := func(s string) time.Time {
			if s == "" {
				return time.Time{}
			}

			formats := []string{
				"2006-01-02T15:04:05",
				time.RFC3339,
				"2006-01-02 15:04:05",
			}
			for _, f := range formats {
				if t, err := time.Parse(f, s); err == nil {
					return t
				}
			}
			return time.Time{}
		}

		sDate := parseUE(startDate)
		eDate := parseUE(endDate)
		rDate := parseUE(regEndDate)

		h := entity.Hackathon{
			ID:              fmt.Sprintf("wue-%d", int(idVal)),
			Title:           title,
			Platform:        "whereuelevate",
			Mode:            mode,
			Location:        location,
			Country:         "India",
			StartDate:       sDate,
			EndDate:         eDate,
			RegistrationEnd: rDate,
			URL:             fmt.Sprintf("https://whereuelevate.com/drills/%s", slug),
		}

		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	}

	return results, nil
}
