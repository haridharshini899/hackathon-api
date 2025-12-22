package fetcher

import (
	"context"
	"fmt"
	"hackathon-api/pkg/entity"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ReSkillFetcher struct{}

func NewReSkillFetcher() *ReSkillFetcher {
	return &ReSkillFetcher{}
}

func (f *ReSkillFetcher) Name() string {
	return "reskill"
}

func (f *ReSkillFetcher) Fetch(ctx context.Context) ([]entity.Hackathon, error) {

	url := "https://reskilll.com/allhacks"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", GetRandomUserAgent())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reskill: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reskill html: %w", err)
	}

	var results []entity.Hackathon

	doc.Find("a.allhackname").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		link, _ := s.Attr("href")
		if link != "" && !strings.HasPrefix(link, "http") {
			link = "https://reskilll.com" + link
		}

		if title == "" {
			return
		}

		var dateStr string
		s.Parents().EachWithBreak(func(_ int, sel *goquery.Selection) bool {
			if sel.Find(".hackregisterdatehead").Length() > 0 {

				sel.Find(".hackregisterdatehead").Each(func(_ int, head *goquery.Selection) {
					if strings.Contains(head.Text(), "Registration End") {
						dateStr = strings.TrimSpace(head.Next().Text())
					}
				})
				return false
			}
			return true
		})

		var endDate time.Time
		if dateStr != "" {

			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				endDate = t
			}
		}

		h := entity.Hackathon{
			ID:              fmt.Sprintf("reskill-%d", i),
			Title:           title,
			Platform:        "reskill",
			URL:             link,
			Country:         "India",
			Mode:            "online",
			StartDate:       time.Time{},
			EndDate:         time.Time{},
			RegistrationEnd: endDate,
		}

		if h.IsIndiaSpecific() {
			results = append(results, h)
		}
	})

	return results, nil
}
