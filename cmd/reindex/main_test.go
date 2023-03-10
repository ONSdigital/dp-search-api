package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v3/client"
	v710 "github.com/ONSdigital/dp-elasticsearch/v3/client/elasticsearch/v710"
	mocks "github.com/ONSdigital/dp-search-api/clients/mock"
	importerModels "github.com/ONSdigital/dp-search-data-importer/models"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testCollectionID = "testCollectionID"
	testDatasetID    = "TS056"
	testEdition      = "2021"
	testVersion      = "4"
	testIndexName    = "ons"
	testMaxIndexing  = 3
	testAuthToken    = "testAuthToken"
)

var (
	ctx     = context.Background()
	testURI = fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", testDatasetID, testEdition, testVersion)
)

func TestCreateIndexName(t *testing.T) {
	Convey("CreateIndexName returns the an index name with the expected prefix", t, func() {
		s0 := createIndexName(testIndexName)
		So(s0, ShouldStartWith, testIndexName)

		Convey("And calling createIndexName again results in a different name", func() {
			s1 := createIndexName(testIndexName)
			So(s1, ShouldNotEqual, s0)
		})
	})
}

func TestTransformMetadataDoc(t *testing.T) {
	Convey("Given a metadata channel and a transformed document channel", t, func() {
		metadataChan := make(chan DatasetMetadata, 1)
		transformedChan := make(chan Document, 1)

		Convey("When a generic dataset metadata is sent to the channel and consumed by transformMetadataDoc", func() {
			sent := DatasetMetadata{
				metadata: &dataset.Metadata{
					DatasetLinks: dataset.Links{
						LatestVersion: dataset.Link{
							URL: fmt.Sprintf("http://testHost:123%s", testURI),
						},
					},
					DatasetDetails: dataset.DatasetDetails{
						ID: testDatasetID,
					},
				},
				isBasedOn: &dataset.IsBasedOn{
					Type: "testType",
					ID:   "testID",
				},
			}

			expected := &importerModels.EsModel{
				DataType:       "dataset_landing_page",
				URI:            testURI,
				DatasetID:      testDatasetID,
				Edition:        testEdition,
				PopulationType: &importerModels.EsPopulationType{},
			}

			metadataChan <- sent
			close(metadataChan)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func(waitGroup *sync.WaitGroup) {
				transformMetadataDoc(ctx, metadataChan, transformedChan, waitGroup)
			}(wg)

			Convey("Then the expected elasticsearch document is sent to the transformed channel", func() {
				transformed := <-transformedChan
				So(transformed.ID, ShouldEqual, testDatasetID)
				So(transformed.URI, ShouldEqual, testURI)

				esModel := &importerModels.EsModel{}
				err := json.Unmarshal(transformed.Body, esModel)
				So(err, ShouldBeNil)
				So(esModel, ShouldResemble, expected)

				wg.Wait()
			})
		})
	})

	Convey("Given a metadata channel and a transformed document channel", t, func() {
		metadataChan := make(chan DatasetMetadata, 1)
		transformedChan := make(chan Document, 1)

		Convey("When a cantabular type dataset metadata is sent to the channel and consumed by transformMetadataDoc", func() {
			areaTypeTrue := true
			areaTypeFalse := false
			sent := DatasetMetadata{
				metadata: &dataset.Metadata{
					DatasetLinks: dataset.Links{
						LatestVersion: dataset.Link{
							URL: fmt.Sprintf("http://testHost:123%s", testURI),
						},
					},
					DatasetDetails: dataset.DatasetDetails{
						ID: testDatasetID,
					},
					Version: dataset.Version{
						Dimensions: []dataset.VersionDimension{
							{ID: "dim1", Label: "label 1 (10 categories)"},
							{ID: "dim2", Label: "label 2 (12 Categories)", IsAreaType: &areaTypeFalse},
							{ID: "dim3", IsAreaType: &areaTypeTrue},
							{ID: "dim4", Label: "label 4 (1 category)"},
						},
					},
				},
				isBasedOn: &dataset.IsBasedOn{
					ID:   "UR_HH",
					Type: "cantabular_flexible_table",
				},
			}

			expected := &importerModels.EsModel{
				DataType:  "cantabular_flexible_table",
				URI:       testURI,
				DatasetID: testDatasetID,
				Edition:   testEdition,
				PopulationType: &importerModels.EsPopulationType{
					Name:  "UR_HH",
					Label: "All usual residents in households",
				},
				Dimensions: []importerModels.EsDimension{
					{Name: "dim1", RawLabel: "label 1 (10 categories)", Label: "label 1"},
					{Name: "dim2", RawLabel: "label 2 (12 Categories)", Label: "label 2"},
					{Name: "dim4", RawLabel: "label 4 (1 category)", Label: "label 4"},
				},
			}

			metadataChan <- sent
			close(metadataChan)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func(waitGroup *sync.WaitGroup) {
				transformMetadataDoc(ctx, metadataChan, transformedChan, waitGroup)
			}(wg)

			Convey("Then the expected elasticsearch document is sent to the transformed channel", func() {
				transformed := <-transformedChan
				So(transformed.ID, ShouldEqual, testDatasetID)
				So(transformed.URI, ShouldEqual, testURI)

				esModel := &importerModels.EsModel{}
				err := json.Unmarshal(transformed.Body, esModel)
				So(err, ShouldBeNil)
				So(esModel, ShouldResemble, expected)

				wg.Wait()
			})
		})
	})
}

func TestRetrieveDatasetEditions(t *testing.T) {
	testEditionDetails := []dataset.EditionsDetails{
		{
			ID: "editionZero",
			Current: dataset.Edition{
				Edition: testEdition,
				Links: dataset.Links{
					LatestVersion: dataset.Link{
						ID: testVersion,
					},
				},
			},
		},
		{
			Current: dataset.Edition{
				Edition: "shouldBeIgnored",
			},
		},
	}
	testIsBasedOn := dataset.IsBasedOn{
		ID:   "UR_HH",
		Type: "cantabular_flexible_table",
	}

	Convey("Given a dataset client that succeeds to return multiple editions where only one has ID and link, and a datasetChan channel", t, func() {
		cli := &mocks.DatasetAPIClientMock{
			GetFullEditionsDetailsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) ([]dataset.EditionsDetails, error) {
				return testEditionDetails, nil
			},
		}
		datasetChan := make(chan dataset.Dataset, 1)

		Convey("When a valid dataset is sent to the dataset channel and consumed by retrieveDatasetEditions", func() {
			datasetChan <- dataset.Dataset{
				Current: &dataset.DatasetDetails{
					ID:        testDatasetID,
					IsBasedOn: &testIsBasedOn,
				},
				DatasetDetails: dataset.DatasetDetails{
					CollectionID: testCollectionID,
				},
			}
			close(datasetChan)

			editionChan, wg := retrieveDatasetEditions(ctx, cli, datasetChan, testAuthToken)

			Convey("Then the expected editions and isBasedOn are sent to the edition channel returned by retrieveDatasetEditions", func() {
				ed1 := <-editionChan
				So(ed1, ShouldResemble, DatasetEditionMetadata{
					id:        testDatasetID,
					editionID: testEdition,
					version:   testVersion,
					isBasedOn: &testIsBasedOn,
				})

				Convey("And the expected call is performed against dataset api", func() {
					So(cli.GetFullEditionsDetailsCalls(), ShouldHaveLength, 1)
					So(cli.GetFullEditionsDetailsCalls()[0].DatasetID, ShouldEqual, testDatasetID)
					So(cli.GetFullEditionsDetailsCalls()[0].CollectionID, ShouldEqual, testCollectionID)
					So(cli.GetFullEditionsDetailsCalls()[0].ServiceAuthToken, ShouldEqual, testAuthToken)
					wg.Wait()
				})
			})
		})
	})
}

func TestRetrieveLatestMetadata(t *testing.T) {
	testMetadata := dataset.Metadata{
		DatasetLinks: dataset.Links{
			LatestVersion: dataset.Link{
				URL: "latestURL",
			},
		},
	}
	testIsBasedOn := dataset.IsBasedOn{
		ID:   "UR_HH",
		Type: "cantabular_flexible_table",
	}

	Convey("Given a dataset client that succeeds to return valid metadata and an editionMetadata channel", t, func() {
		cli := &mocks.DatasetAPIClientMock{
			GetVersionMetadataFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.Metadata, error) {
				return testMetadata, nil
			},
		}
		editionMetadata := make(chan DatasetEditionMetadata, 1)

		Convey("When a dataset edition metadata is sent to the edition metadata channel and consumed by retrieveLatestMetadata", func() {
			editionMetadata <- DatasetEditionMetadata{
				id:        testDatasetID,
				editionID: testEdition,
				version:   testVersion,
				isBasedOn: &testIsBasedOn,
			}
			close(editionMetadata)

			metadataChan, wg := retrieveLatestMetadata(ctx, cli, editionMetadata, testAuthToken)

			Convey("Then the expected metadata and isBasedOn are sent to the metadataChannel", func() {
				m := <-metadataChan
				So(m.metadata, ShouldResemble, &testMetadata)
				So(m.isBasedOn, ShouldResemble, &testIsBasedOn)

				Convey("And the expected call is performed against dataset api", func() {
					So(cli.GetVersionMetadataCalls(), ShouldHaveLength, 1)
					So(cli.GetVersionMetadataCalls()[0].ID, ShouldEqual, testDatasetID)
					So(cli.GetVersionMetadataCalls()[0].Edition, ShouldEqual, testEdition)
					So(cli.GetVersionMetadataCalls()[0].Version, ShouldEqual, testVersion)
					So(cli.GetVersionMetadataCalls()[0].ServiceAuthToken, ShouldEqual, testAuthToken)
					wg.Wait()
				})
			})
		})
	})
}

func TestIndexDoc(t *testing.T) {
	Convey("Given a successful elasticsearch client mock, a Document channel and an indexed bool channel", t, func() {
		esClient := &mocks.ElasticSearchMock{
			BulkIndexAddFunc: func(
				ctx context.Context,
				action dpEsClient.BulkIndexerAction,
				index string,
				documentID string,
				document []byte,
				onSuccess dpEsClient.SuccessFunc,
				onFailure dpEsClient.FailureFunc,
			) error {
				onSuccess(ctx, esutil.BulkIndexerItem{}, esutil.BulkIndexerResponseItem{})
				return nil
			},
		}
		transformedChan := make(chan Document, 1)
		indexedChan := make(chan bool)

		Convey("When a Document is sent to the transformedChan and consumed by indexDoc", func() {
			transformedChan <- Document{
				ID:   testDatasetID,
				URI:  testURI,
				Body: []byte{1, 2, 3, 4},
			}
			close(transformedChan)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				indexDoc(ctx, esClient, transformedChan, indexedChan, testIndexName)
			}()

			Convey("Then the document is indexed and the expected bulkAdd call is performed", func() {
				indexed := <-indexedChan
				So(indexed, ShouldBeTrue)

				So(esClient.BulkIndexAddCalls(), ShouldHaveLength, 1)
				So(esClient.BulkIndexAddCalls()[0].Action, ShouldEqual, v710.Create)
				So(esClient.BulkIndexAddCalls()[0].Document, ShouldResemble, []byte{1, 2, 3, 4})
				So(esClient.BulkIndexAddCalls()[0].DocumentID, ShouldEqual, testDatasetID)
				So(esClient.BulkIndexAddCalls()[0].Index, ShouldEqual, testIndexName)

				wg.Wait()
			})
		})
	})

	Convey("Given an elasticsearch client mock that returns an error, a Document channel and an indexed bool channel", t, func() {
		esClient := &mocks.ElasticSearchMock{
			BulkIndexAddFunc: func(
				ctx context.Context,
				action dpEsClient.BulkIndexerAction,
				index string,
				documentID string,
				document []byte,
				onSuccess dpEsClient.SuccessFunc,
				onFailure dpEsClient.FailureFunc,
			) error {
				return errors.New("testError")
			},
		}
		transformedChan := make(chan Document, 1)
		indexedChan := make(chan bool)

		Convey("When a Document is sent to the transformedChan and consumed by indexDoc", func() {
			transformedChan <- Document{
				ID:   testDatasetID,
				URI:  testURI,
				Body: []byte{1, 2, 3, 4},
			}
			close(transformedChan)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				indexDoc(ctx, esClient, transformedChan, indexedChan, testIndexName)
			}()

			Convey("Then the document is not indexed and the expected bulkAdd call is performed", func() {
				indexed := <-indexedChan
				So(indexed, ShouldBeFalse)

				So(esClient.BulkIndexAddCalls(), ShouldHaveLength, 1)
				So(esClient.BulkIndexAddCalls()[0].Action, ShouldEqual, v710.Create)
				So(esClient.BulkIndexAddCalls()[0].Document, ShouldResemble, []byte{1, 2, 3, 4})
				So(esClient.BulkIndexAddCalls()[0].DocumentID, ShouldEqual, testDatasetID)
				So(esClient.BulkIndexAddCalls()[0].Index, ShouldEqual, testIndexName)

				wg.Wait()
			})
		})
	})

	Convey("Given an elasticsearch client mock that fails asynchronously, a Document channel and an indexed bool channel", t, func() {
		esClient := &mocks.ElasticSearchMock{
			BulkIndexAddFunc: func(
				ctx context.Context,
				action dpEsClient.BulkIndexerAction,
				index string,
				documentID string,
				document []byte,
				onSuccess dpEsClient.SuccessFunc,
				onFailure dpEsClient.FailureFunc,
			) error {
				onFailure(ctx, esutil.BulkIndexerItem{}, esutil.BulkIndexerResponseItem{}, errors.New("testError"))
				return nil
			},
		}
		transformedChan := make(chan Document, 1)
		indexedChan := make(chan bool)

		Convey("When a Document is sent to the transformedChan and consumed by indexDoc", func() {
			transformedChan <- Document{
				ID:   testDatasetID,
				URI:  testURI,
				Body: []byte{1, 2, 3, 4},
			}
			close(transformedChan)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				indexDoc(ctx, esClient, transformedChan, indexedChan, testIndexName)
			}()

			Convey("Then the document is not indexed and the expected bulkAdd call is performed", func() {
				indexed := <-indexedChan
				So(indexed, ShouldBeFalse)

				So(esClient.BulkIndexAddCalls(), ShouldHaveLength, 1)
				So(esClient.BulkIndexAddCalls()[0].Action, ShouldEqual, v710.Create)
				So(esClient.BulkIndexAddCalls()[0].Document, ShouldResemble, []byte{1, 2, 3, 4})
				So(esClient.BulkIndexAddCalls()[0].DocumentID, ShouldEqual, testDatasetID)
				So(esClient.BulkIndexAddCalls()[0].Index, ShouldEqual, testIndexName)

				wg.Wait()
			})
		})
	})
}
