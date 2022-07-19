package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	dpEs "github.com/ONSdigital/dp-elasticsearch/v3"
	dpEsClient "github.com/ONSdigital/dp-elasticsearch/v3/client"
	v710 "github.com/ONSdigital/dp-elasticsearch/v3/client/elasticsearch/v710"
	"github.com/ONSdigital/dp-net/v2/awsauth"
	dphttp2 "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-search-api/clients"
	"github.com/ONSdigital/dp-search-api/elasticsearch"
	extractorModels "github.com/ONSdigital/dp-search-data-extractor/models"
	importerModels "github.com/ONSdigital/dp-search-data-importer/models"
	"github.com/ONSdigital/dp-search-data-importer/transform"
)

var (
	maxConcurrentExtractions = 20
	maxConcurrentIndexings   = 30
	DefaultPaginationLimit   = 500
)

type cliConfig struct {
	aws              AWSConfig
	zebedeeURL       string
	datasetURL       string
	esURL            string
	signRequests     bool
	ServiceAuthToken string
}

type AWSConfig struct {
	filename              string
	profile               string
	region                string
	service               string
	tlsInsecureSkipVerify bool
}

type DatasetEditionMetadata struct {
	id        string
	editionID string
	version   string
}

type Document struct {
	ID   string
	URI  string
	Body []byte
}

func main() {
	fmt.Printf("Hola %s!\n", Name)

	ctx := context.Background()
	cfg := getConfig(ctx)

	hcClienter := dphttp2.NewClient()
	if hcClienter == nil {
		log.Fatal("failed to create dp http client")
	}
	hcClienter.SetMaxRetries(2)
	hcClienter.SetTimeout(30 * time.Second) // Published Index takes about 10s to return so add a bit more
	zebClient := zebedee.NewClientWithClienter(cfg.zebedeeURL, hcClienter)
	if zebClient == nil {
		log.Fatal("failed to create zebedee client")
	}

	esHTTPClient := hcClienter
	if cfg.signRequests {
		fmt.Println("Use a signing roundtripper client")
		awsSignerRT, err := awsauth.NewAWSSignerRoundTripper(cfg.aws.filename, cfg.aws.filename, cfg.aws.region, cfg.aws.service,
			awsauth.Options{TlsInsecureSkipVerify: cfg.aws.tlsInsecureSkipVerify})
		if err != nil {
			log.Fatal(ctx, "Failed to create http signer", err)
		}

		esHTTPClient = dphttp2.NewClientWithTransport(awsSignerRT)
	}

	datasetClient := dataset.NewAPIClient(cfg.datasetURL)
	esClient, esClientErr := dpEs.NewClient(dpEsClient.Config{
		ClientLib: dpEsClient.GoElasticV710,
		Address:   cfg.esURL,
		Transport: esHTTPClient,
	})
	if esClientErr != nil {
		log.Fatal(ctx, "Failed to create dp-elasticsearch client", esClientErr)
	}

	if err := esClient.NewBulkIndexer(ctx); err != nil {
		log.Fatal(ctx, "Failed to create new bulk indexer")
	}
	datasetChan := extractDatasets(ctx, datasetClient, cfg.ServiceAuthToken)
	editionChan := retrieveDatasetEditions(ctx, datasetClient, datasetChan, cfg.ServiceAuthToken)
	metadataChan := retrieveLatestMetadata(ctx, datasetClient, editionChan, cfg.ServiceAuthToken)
	urisChan := uriProducer(ctx, zebClient)
	extractedChan, extractionFailuresChan := docExtractor(ctx, zebClient, urisChan, maxConcurrentExtractions)
	transformedChan := docTransformer(extractedChan, metadataChan)
	indexedChan := docIndexer(ctx, esClient, transformedChan, maxConcurrentIndexings)

	summarize(indexedChan, extractionFailuresChan)
	cleanOldIndices(ctx, esClient)
}

func uriProducer(ctx context.Context, z clients.ZebedeeClient) chan string {
	uriChan := make(chan string)
	go func() {
		defer close(uriChan)
		items := getPublishedURIs(ctx, z)
		for _, item := range items {
			uriChan <- item.URI
		}
		fmt.Println("Finished listing uris")
	}()
	return uriChan
}

func getPublishedURIs(ctx context.Context, z clients.ZebedeeClient) []zebedee.PublishedIndexItem {
	index, err := z.GetPublishedIndex(ctx, &zebedee.PublishedIndexRequestParams{})
	if err != nil {
		log.Fatalf("Fatal error getting index from zebedee: %s", err)
	}
	fmt.Printf("Fetched %d uris from zebedee\n", index.Count)
	return index.Items
}

func docExtractor(ctx context.Context, z clients.ZebedeeClient, uriChan chan string, maxExtractions int) (extractedChan chan Document, extractionFailuresChan chan string) {
	extractedChan = make(chan Document)
	extractionFailuresChan = make(chan string)
	go func() {
		defer close(extractedChan)
		defer close(extractionFailuresChan)

		var wg sync.WaitGroup

		for w := 0; w < maxExtractions; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				extractDoc(ctx, z, uriChan, extractedChan, extractionFailuresChan)
			}()
		}
		wg.Wait()
		fmt.Println("Finished extracting docs")
	}()
	return
}

func extractDoc(ctx context.Context, z clients.ZebedeeClient, uriChan <-chan string, extractedChan chan Document, extractionFailuresChan chan string) {
	for uri := range uriChan {
		body, err := z.GetPublishedData(ctx, uri)
		if err != nil {
			extractionFailuresChan <- uri
		}

		extractedDoc := Document{
			URI:  uri,
			Body: body,
		}
		extractedChan <- extractedDoc
	}
}

func docTransformer(extractedChan chan Document, metadataChan chan dataset.Metadata) chan Document {
	transformedChan := make(chan Document)
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < maxConcurrentExtractions; i++ {
			wg.Add(2)
			go func(wg *sync.WaitGroup) {
				transformZebedeeDoc(extractedChan, transformedChan, wg)
				transformMetadataDoc(metadataChan, transformedChan, wg)
			}(&wg)
		}
		wg.Wait()
		close(transformedChan)
		fmt.Println("Finished transforming docs")
	}()
	return transformedChan
}

func transformZebedeeDoc(extractedChan chan Document, transformedChan chan<- Document, wg *sync.WaitGroup) {
	defer wg.Done()
	var wg2 sync.WaitGroup
	for extractedDoc := range extractedChan {
		wg2.Add(1)
		go func(extractedDoc Document) {
			defer wg2.Done()
			var zebedeeData extractorModels.ZebedeeData
			err := json.Unmarshal(extractedDoc.Body, &zebedeeData)
			if err != nil {
				log.Fatal("error while attempting to unmarshal zebedee response into zebedeeData", err) // TODO proper error handling
			}
			exporterEventData := extractorModels.MapZebedeeDataToSearchDataImport(zebedeeData, -1)
			importerEventData := convertToSearchDataModel(exporterEventData)
			esModel := transform.NewTransformer().TransformEventModelToEsModel(&importerEventData)

			body, err := json.Marshal(esModel)
			if err != nil {
				log.Fatal("error marshal to json", err) // TODO error handling
				return
			}

			transformedDoc := Document{
				ID:   exporterEventData.UID,
				URI:  extractedDoc.URI,
				Body: body,
			}
			transformedChan <- transformedDoc
		}(extractedDoc)
	}
	wg2.Wait()
}

func transformMetadataDoc(metadataChan chan dataset.Metadata, transformedChan chan<- Document, wg *sync.WaitGroup) {
	for metadata := range metadataChan {
		uri := metadata.DatasetLinks.LatestVersion.URL
		parsedURI, err := url.Parse(uri)
		if err != nil {
			log.Fatalf("error occured while parsing url: %v", err)
		}
		cmdData := extractorModels.CMDData{
			UID: metadata.DatasetDetails.ID,
			URI: parsedURI.Path,
			VersionDetails: extractorModels.VersionDetails{
				ReleaseDate: metadata.Version.ReleaseDate,
			},
			DatasetDetails: extractorModels.DatasetDetails{
				Title:       metadata.DatasetDetails.Title,
				Description: metadata.DatasetDetails.Description,
			},
		}
		if metadata.DatasetDetails.Keywords != nil {
			cmdData.DatasetDetails.Keywords = *metadata.DatasetDetails.Keywords
		}
		exporterEventData := extractorModels.MapVersionMetadataToSearchDataImport(cmdData)
		importerEventData := convertToSearchDataModel(exporterEventData)
		esModel := transform.NewTransformer().TransformEventModelToEsModel(&importerEventData)
		body, err := json.Marshal(esModel)
		if err != nil {
			wg.Done()
			log.Fatal("error marshal to json", err) // TODO error handling
			return
		}

		transformedDoc := Document{
			ID:   exporterEventData.UID,
			URI:  parsedURI.Path,
			Body: body,
		}
		transformedChan <- transformedDoc
	}
	wg.Done()
}

func docIndexer(ctx context.Context, dpEsIndexClient dpEsClient.Client, transformedChan chan Document, maxIndexings int) chan bool {
	indexedChan := make(chan bool)
	go func() {
		defer close(indexedChan)

		indexName := createIndexName("ons")

		err := dpEsIndexClient.CreateIndex(ctx, indexName, elasticsearch.GetSearchIndexSettings())
		if err != nil {
			log.Fatal(ctx, "error creating index", err)
		}

		fmt.Printf("Index created: %s\n", indexName)

		var wg sync.WaitGroup

		for w := 0; w < maxIndexings; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				indexDoc(ctx, dpEsIndexClient, transformedChan, indexedChan, indexName)
			}()
		}
		wg.Wait()
		dpEsIndexClient.BulkIndexClose(ctx)
		fmt.Println("Finished indexing docs")

		swapAliases(ctx, dpEsIndexClient, indexName)
	}()
	return indexedChan
}

func createIndexName(s string) string {
	now := time.Now()
	return fmt.Sprintf("%s%d", s, now.UnixMicro())
}

func indexDoc(ctx context.Context, esClient dpEsClient.Client, transformedChan <-chan Document, indexedChan chan bool, indexName string) {
	for transformedDoc := range transformedChan {
		indexed := true
		err := esClient.BulkIndexAdd(ctx, v710.Create, indexName, transformedDoc.ID, transformedDoc.Body)
		if err != nil {
			indexed = false
		}

		indexedChan <- indexed
	}
}

func swapAliases(ctx context.Context, dpEsIndexClient dpEsClient.Client, indexName string) {
	updateAliasErr := dpEsIndexClient.UpdateAliases(ctx, "ons", []string{"ons*"}, []string{indexName})
	if updateAliasErr != nil {
		log.Fatalf("error swapping aliases: %v", updateAliasErr)
	}
}

func summarize(indexedChan <-chan bool, extractionFailuresChan <-chan string) {
	totalIndexed, totalFailed := 0, 0
	for range extractionFailuresChan {
		totalFailed++
	}
	for indexed := range indexedChan {
		if indexed {
			totalIndexed++
		} else {
			totalFailed++
		}
	}
	fmt.Printf("Indexed: %d, Failed: %d\n", totalIndexed, totalFailed)
}

type aliasResponse map[string]indexDetails

type indexDetails struct {
	Aliases map[string]interface{} `json:"aliases"`
}

func cleanOldIndices(ctx context.Context, dpEsIndexClient dpEsClient.Client) {
	body, err := dpEsIndexClient.GetAlias(ctx) // Create this method via dp-elasticsearch v3 lib
	if err != nil {
		log.Fatalf("Error: Indices.GetAlias: %s", err)
	}
	var r aliasResponse
	if err := json.Unmarshal(body, &r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	toDelete := []string{}
	for index, details := range r {
		if strings.HasPrefix(index, "ons") && !doesIndexHaveAlias(details, "ons") {
			toDelete = append(toDelete, index)
		}
	}

	deleteIndicies(ctx, dpEsIndexClient, toDelete)
}

func doesIndexHaveAlias(details indexDetails, alias string) bool {
	for k := range details.Aliases {
		if k == alias {
			return true
		}
	}
	return false
}

func deleteIndicies(ctx context.Context, dpEsIndexClient dpEsClient.Client, indicies []string) {
	if err := dpEsIndexClient.DeleteIndices(ctx, indicies); err != nil {
		log.Fatalf("Error: Indices.GetAlias: %s", err)
	}
	fmt.Printf("Deleted Indicies: %s\n", strings.Join(indicies, ","))
}

func extractDatasets(ctx context.Context, datasetClient clients.DatasetAPIClient, serviceAuthToken string) chan dataset.Dataset {
	datasetChan := make(chan dataset.Dataset)
	go func() {
		defer close(datasetChan)
		var list dataset.List
		var err error
		var offset = 0
		for {
			list, err = datasetClient.GetDatasets(ctx, "", serviceAuthToken, "", &dataset.QueryParams{
				Offset: offset,
				Limit:  DefaultPaginationLimit,
			})
			if err != nil {
				log.Fatalf("Error: retrieving dataset clients: %v", err)
			}
			if len(list.Items) == 0 {
				break
			}
			for i := 0; i < len(list.Items); i++ {
				datasetChan <- list.Items[i]
			}
			offset += DefaultPaginationLimit
		}
	}()
	return datasetChan
}

func retrieveDatasetEditions(ctx context.Context, datasetClient clients.DatasetAPIClient, datasetChan chan dataset.Dataset, serviceAuthToken string) chan DatasetEditionMetadata {
	editionMetadataChan := make(chan DatasetEditionMetadata)
	var wg sync.WaitGroup
	go func() {
		defer close(editionMetadataChan)
		for i := 0; i < maxConcurrentExtractions; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for dataset := range datasetChan {
					if dataset.Current == nil {
						continue
					}
					editions, err := datasetClient.GetFullEditionsDetails(ctx, "", serviceAuthToken, dataset.CollectionID, dataset.Current.ID)
					if err != nil {
						log.Printf("error retrieving editions with dataset id: %v", dataset.ID)
					} else {
						for i := 0; i < len(editions); i++ {
							if editions[i].ID == "" || editions[i].Current.Links.LatestVersion.ID == "" {
								continue
							}
							editionMetadataChan <- DatasetEditionMetadata{
								id:        dataset.Current.ID,
								editionID: editions[i].Current.Edition,
								version:   editions[i].Current.Links.LatestVersion.ID,
							}
						}
					}
				}
			}()
		}
		wg.Wait()
	}()
	return editionMetadataChan
}

func retrieveLatestMetadata(ctx context.Context, datasetClient clients.DatasetAPIClient, editionMetadata chan DatasetEditionMetadata, serviceAuthToken string) chan dataset.Metadata {
	metadataChan := make(chan dataset.Metadata)
	var wg sync.WaitGroup
	go func() {
		defer close(metadataChan)
		for i := 0; i < maxConcurrentExtractions; i++ {
			wg.Add(1)
			go func() {
				for edMetadata := range editionMetadata {
					metadata, err := datasetClient.GetVersionMetadata(ctx, "", serviceAuthToken, "", edMetadata.id, edMetadata.editionID, edMetadata.version)
					if err != nil {
						continue
					}
					metadataChan <- metadata
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}()
	return metadataChan
}

func convertToSearchDataModel(searchDataImport extractorModels.SearchDataImport) importerModels.SearchDataImportModel {
	searchDIM := importerModels.SearchDataImportModel{
		UID:             searchDataImport.UID,
		URI:             searchDataImport.URI,
		DataType:        searchDataImport.DataType,
		JobID:           searchDataImport.JobID,
		SearchIndex:     searchDataImport.SearchIndex,
		CanonicalTopic:  searchDataImport.CanonicalTopic,
		CDID:            searchDataImport.CDID,
		DatasetID:       searchDataImport.DatasetID,
		Keywords:        searchDataImport.Keywords,
		MetaDescription: searchDataImport.MetaDescription,
		ReleaseDate:     searchDataImport.ReleaseDate,
		Summary:         searchDataImport.Summary,
		Title:           searchDataImport.Title,
		Topics:          searchDataImport.Topics,
		TraceID:         searchDataImport.TraceID,
		Cancelled:       searchDataImport.Cancelled,
		Finalised:       searchDataImport.Finalised,
		ProvisionalDate: searchDataImport.ProvisionalDate,
		Published:       searchDataImport.Published,
		Survey:          searchDataImport.Survey,
		Language:        searchDataImport.Language,
	}
	for _, dateChange := range searchDataImport.DateChanges {
		searchDIM.DateChanges = append(searchDIM.DateChanges, importerModels.ReleaseDateDetails{
			ChangeNotice: dateChange.ChangeNotice,
			Date:         dateChange.Date,
		})
	}
	return searchDIM
}
