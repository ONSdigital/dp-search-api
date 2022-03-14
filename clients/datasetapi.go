package clients

import (
	"context"

	datasetclient "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
)

//go:generate moq -out mock/datasetapi.go -pkg mock . DatasetClient

// DatasetAPIClient defines the zebedee client
type DatasetAPIClient interface {
	GetDatasets(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, q *datasetclient.QueryParams) (m datasetclient.List, err error)
	GetEditions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m []datasetclient.Edition, err error)
	GetFullEditionsDetails(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m []datasetclient.EditionsDetails, err error)
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m datasetclient.Metadata, err error)
}
