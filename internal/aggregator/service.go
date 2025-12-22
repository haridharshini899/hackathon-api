package aggregator

import (
	"context"
	"hackathon-api/internal/entity"
	"hackathon-api/internal/fetcher"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type Service struct {
	fetchers []fetcher.Fetcher
	cache    *cache.Cache
}

func NewService(fetchers []fetcher.Fetcher) *Service {

	c := cache.New(15*time.Minute, 20*time.Minute)
	return &Service{
		fetchers: fetchers,
		cache:    c,
	}
}

func (s *Service) GetHackathons(ctx context.Context, mode, platform string) ([]entity.Hackathon, error) {

	cacheKey := "all_hackathons"
	if found, ok := s.cache.Get(cacheKey); ok {

		allData := found.([]entity.Hackathon)
		return filterAndSort(allData, mode, platform), nil
	}

	var wg sync.WaitGroup
	resultChan := make(chan []entity.Hackathon, len(s.fetchers))

	for _, f := range s.fetchers {
		wg.Add(1)
		go func(f fetcher.Fetcher) {
			defer wg.Done()

			fCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
			defer cancel()

			items, err := f.Fetch(fCtx)
			if err != nil {

				return
			}
			resultChan <- items
		}(f)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var allHackathons []entity.Hackathon
	for items := range resultChan {
		allHackathons = append(allHackathons, items...)
	}

	unique := deduplicate(allHackathons)

	s.cache.Set(cacheKey, unique, cache.DefaultExpiration)

	return filterAndSort(unique, mode, platform), nil
}

func deduplicate(items []entity.Hackathon) []entity.Hackathon {
	seen := make(map[string]bool)
	var result []entity.Hackathon

	for _, h := range items {

		key := strings.ToLower(h.Title)

		if !seen[key] {
			seen[key] = true
			result = append(result, h)
		}
	}
	return result
}

func filterAndSort(items []entity.Hackathon, mode, platform string) []entity.Hackathon {
	var filtered []entity.Hackathon
	now := time.Now().Add(-24 * time.Hour)

	for _, h := range items {

		if !h.EndDate.IsZero() && h.EndDate.Before(now) {
			continue
		}

		if !h.RegistrationEnd.IsZero() && h.RegistrationEnd.Before(now) {
			continue
		}

		if platform != "" {

			match := false

			inputPlatforms := strings.Split(platform, ",")
			for _, p := range inputPlatforms {
				if strings.EqualFold(strings.TrimSpace(p), h.Platform) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		if mode != "" && !strings.EqualFold(h.Mode, mode) {
			continue
		}

		filtered = append(filtered, h)
	}

	var valid []entity.Hackathon
	for _, h := range filtered {
		if h.StartDate.IsZero() && h.EndDate.IsZero() && h.RegistrationEnd.IsZero() {
			continue
		}
		valid = append(valid, h)
	}

	sort.SliceStable(valid, func(i, j int) bool {

		getDate := func(h entity.Hackathon) time.Time {
			if !h.StartDate.IsZero() {
				return h.StartDate
			}
			if !h.EndDate.IsZero() {
				return h.EndDate
			}
			return h.RegistrationEnd
		}

		d1 := getDate(valid[i])
		d2 := getDate(valid[j])

		if d1.Equal(d2) {
			return strings.ToLower(valid[i].Title) < strings.ToLower(valid[j].Title)
		}
		return d1.Before(d2)
	})

	return valid
}
