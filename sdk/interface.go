package sdk

import (
	"context"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-api/models"
	apiError "github.com/ONSdigital/dp-search-api/sdk/errors"
	"github.com/ONSdigital/dp-search-api/transformer"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Clienter

type Clienter interface {
	Checker(ctx context.Context, check *health.CheckState) error
	CreateIndex(ctx context.Context, options Options) (*models.CreateIndexResponse, apiError.Error)
	GetReleaseCalendarEntries(ctx context.Context, options Options) (*transformer.ReleaseTransformer, apiError.Error)
	GetSearch(ctx context.Context, options Options) (*models.SearchResponse, apiError.Error)
	Health() *healthcheck.Client
	URL() string
}
