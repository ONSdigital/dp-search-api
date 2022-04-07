package clients

import (
	"context"

	datasetclient "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
)

//go:generate moq -out mock/datasetapi.go -pkg mock . DatasetClient

// DatasetAPIClient defines the dataset client
type DatasetAPIClient interface {
	GetDatasets(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, q *datasetclient.QueryParams) (m datasetclient.List, err error)
	GetEditions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m []datasetclient.Edition, err error)
	GetFullEditionsDetails(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m []datasetclient.EditionsDetails, err error)
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m datasetclient.Metadata, err error)
}

// ZebedeeClient defines the zebedee client
type ZebedeeClient interface {
	GetPublishedIndex(ctx context.Context, piRequest *zebedee.PublishedIndexRequestParams) (zebedee.PublishedIndex, error)
	GetPublishedData(ctx context.Context, uriString string) ([]byte, error)
}
