package fetcher

import (
	"context"
	"hackathon-api/pkg/entity"
)

type Fetcher interface {
	Name() string

	Fetch(ctx context.Context) ([]entity.Hackathon, error)
}
