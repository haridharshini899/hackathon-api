package main

import (
	"context"
	"fmt"
	"hackathon-api/pkg/fetcher"
	"time"
)

func main() {
	var fetchers = []fetcher.Fetcher{
		fetcher.NewDevfolioFetcher(),
		fetcher.NewUnstopFetcher(),
		fetcher.NewDevpostFetcher(),
		fetcher.NewMLHFetcher(),
		fetcher.NewHackerEarthFetcher(),
		fetcher.NewHack2SkillFetcher(),
		fetcher.NewReSkillFetcher(),
		fetcher.NewWhereUElevateFetcher(),
		fetcher.NewDevnovateFetcher(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	for _, f := range fetchers {
		fmt.Printf("\n=== Testing Fetcher: %s ===\n", f.Name())
		start := time.Now()
		hackathons, err := f.Fetch(ctx)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}

		fmt.Printf("Success! Found %d hackathons in %v\n", len(hackathons), duration)
		if len(hackathons) > 0 {
			v := hackathons[0]
			fmt.Printf("Sample: %s | %s | %v | %s\n\n", v.Title, v.Mode, v.StartDate, v.URL)
		} else {
			fmt.Println("WARNING: 0 results found. Selectors might be broken.")
		}
	}
}
