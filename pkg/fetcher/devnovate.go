package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"hackathon-api/pkg/entity"
	"net/http"
	"strings"
	"time"
)

type DevnovateFetcher struct{}

func NewDevnovateFetcher() *DevnovateFetcher {
	return &DevnovateFetcher{}
}

func (f *DevnovateFetcher) Name() string {
	return "devnovate"
}

func (f *DevnovateFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {
	url := "https://devnovate.co/api/v1/events"

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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")

	tr := &http.Transport{
		ForceAttemptHTTP2:   false,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,

		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	var resp *http.Response
	var reqErr error
	for i := 0; i < 3; i++ {

		resp, reqErr = client.Do(req.WithContext(ctx))
		if reqErr == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	if reqErr != nil {
		return nil, fmt.Errorf("failed to fetch devnovate after retries: %w", reqErr)
	}
	if resp == nil {
		return nil, fmt.Errorf("devnovate failed unknown")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("devnovate returned status: %d", resp.StatusCode)
	}

	var raw interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode devnovate json: %w", err)
	}

	var eventList []interface{}

	if val, ok := raw.([]interface{}); ok {
		eventList = val
	} else if m, ok := raw.(map[string]interface{}); ok {

		if val, ok := m["data"].([]interface{}); ok {
			eventList = val
		} else if val, ok := m["events"].([]interface{}); ok {
			eventList = val
		} else if val, ok := m["results"].([]interface{}); ok {
			eventList = val
		}
	}

	var results []entity.Hackathon
	for _, e := range eventList {
		item, ok := e.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := item["name"].(string)
		if title == "" {
			title, _ = item["eventName"].(string)
		}

		slug, _ := item["hackathon"].(string)
		if slug == "" {
			slug, _ = item["eventName"].(string)
		}

		status, _ := item["status"].(string)
		location, _ := item["location"].(string)

		mode := "offline"
		if strings.EqualFold(status, "online") || strings.EqualFold(location, "online") {
			mode = "online"
		}

		startDateStr, _ := item["startDate"].(string)
		endDateStr, _ := item["endDate"].(string)
		regDeadlineStr, _ := item["registrationDeadline"].(string)

		parseDate := func(s string) time.Time {
			s = strings.TrimSpace(s)
			if s == "" {
				return time.Time{}
			}

			formats := []string{
				"02-01-2006",
				"2006-01-02",
				"2006-01-02T15:04",
				"2006-01-02T15:04:05",
				time.RFC3339,
			}
			for _, f := range formats {
				if t, err := time.Parse(f, s); err == nil {
					return t
				}
			}

			if len(s) > 10 {
				short := s[:10]
				if t, err := time.Parse("2006-01-02", short); err == nil {
					return t
				}
			}

			return time.Time{}
		}

		sDate := parseDate(startDateStr)
		eDate := parseDate(endDateStr)
		regEnd := parseDate(regDeadlineStr)

		idRaw := item["_id"]

		eventUrl := "https://devnovate.co/events"
		if slug != "" {
			eventUrl = fmt.Sprintf("https://devnovate.co/event/%s", slug)
		}

		h := entity.Hackathon{
			ID:              fmt.Sprintf("devnovate-%v", idRaw),
			Title:           title,
			Platform:        "devnovate",
			Mode:            mode,
			Location:        location,
			Country:         "India",
			StartDate:       sDate,
			EndDate:         eDate,
			RegistrationEnd: regEnd,
			URL:             eventUrl,
		}

		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	}

	return results, nil
}
