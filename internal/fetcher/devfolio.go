package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"hackathon-api/internal/entity"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type DevfolioFetcher struct{}

func NewDevfolioFetcher() *DevfolioFetcher {
	return &DevfolioFetcher{}
}

func (f *DevfolioFetcher) Name() string {
	return "devfolio"
}

func (f *DevfolioFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {
	url := "https://devfolio.co/hackathons"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", GetRandomUserAgent())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devfolio: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse devfolio html: %w", err)
	}

	var results []entity.Hackathon
	foundData := false

	doc.Find("script#__NEXT_DATA__").Each(func(i int, s *goquery.Selection) {
		if foundData {
			return
		}
		jsonStr := s.Text()

		var nextData struct {
			Props struct {
				PageProps struct {
					DehydratedState struct {
						Queries []struct {
							State struct {
								Data struct {
									OpenHackathons []struct {
										Name     string `json:"name"`
										Slug     string `json:"slug"`
										Location string `json:"location"`
										StartsAt string `json:"starts_at"`
										EndsAt   string `json:"ends_at"`
										IsOnline bool   `json:"is_online"`
									} `json:"open_hackathons"`
								} `json:"data"`
							} `json:"state"`
						} `json:"queries"`
					} `json:"dehydratedState"`
				} `json:"pageProps"`
			} `json:"props"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &nextData); err == nil {

			for _, query := range nextData.Props.PageProps.DehydratedState.Queries {

				if len(query.State.Data.OpenHackathons) > 0 {
					foundData = true
					for _, h := range query.State.Data.OpenHackathons {
						start, _ := time.Parse(time.RFC3339, h.StartsAt)
						end, _ := time.Parse(time.RFC3339, h.EndsAt)

						mode := "offline"
						if h.IsOnline {
							mode = "online"
						}

						item := entity.Hackathon{
							ID:        fmt.Sprintf("devfolio-%s", h.Slug),
							Title:     h.Name,
							Platform:  "devfolio",
							Mode:      mode,
							Location:  h.Location,
							Country:   "India",
							StartDate: start,
							EndDate:   end,
							URL:       fmt.Sprintf("https://%s.devfolio.co", h.Slug),
						}

						if h.IsOnline {
							item.Country = "India"
						}

						if item.IsIndiaSpecific() {
							results = append(results, item)
						}
					}
				}
			}
		}
	})

	if !foundData {

		doc.Find("h3").Each(func(i int, s *goquery.Selection) {
			title := strings.TrimSpace(s.Text())
			parentLink := s.ParentsFiltered("a").First()
			link, exists := parentLink.Attr("href")
			if exists {
				results = append(results, entity.Hackathon{
					ID:        fmt.Sprintf("devfolio-scrape-%d", i),
					Title:     title,
					Platform:  "devfolio",
					URL:       link,
					Country:   "India",
					Mode:      "online",
					StartDate: time.Time{},
				})
			}
		})
	}

	return results, nil
}
