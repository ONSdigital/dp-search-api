package transformer

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/models"
	"github.com/pkg/errors"
)

// LegacyTransformer represents an instance of the ResponseTransformer interface
type LegacyTransformer struct {
	higlightReplacer *strings.Replacer
}

// Transformer represents an instance of the ResponseTransformer interface for ES7x
type Transformer struct {
	higlightReplacer *strings.Replacer
}

// New returns a new instance of Transformer ES7x
func New(esVersion710 bool) api.ResponseTransformer {
	highlightReplacer := strings.NewReplacer("<em class=\"highlight\">", "", "</em>", "")
	if esVersion710 {
		return &Transformer{
			higlightReplacer: highlightReplacer,
		}
	}
	return &LegacyTransformer{
		higlightReplacer: highlightReplacer,
	}
}

// TransformSearchResponse transforms an elastic search response into a structure that matches the v1 api specification
func (t *LegacyTransformer) TransformSearchResponse(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error) {
	var source models.ESResponseLegacy

	err := json.Unmarshal(responseData, &source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode elastic search response")
	}

	if len(source.Responses) < 1 {
		return nil, errors.New("Response to be transformed contained 0 items")
	}

	sr := t.legayTransform(&source, highlight)

	needAdditionalSuggestions := numberOfSearchTerms(query)
	if needAdditionalSuggestions > 1 {
		as := buildAdditionalSuggestionList(query)
		sr.AdditionSuggestions = as
	}

	transformedData, err := json.Marshal(sr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode transformed response")
	}
	return transformedData, nil
}

func (t *LegacyTransformer) legayTransform(source *models.ESResponseLegacy, highlight bool) models.SearchResponseLegacy {
	sr := models.SearchResponseLegacy{
		Count:        source.Responses[0].Hits.Total,
		Items:        []models.ContentItemLegacy{},
		ContentTypes: []models.ContentType{},
	}
	var took int
	for _, response := range source.Responses {
		for i := 0; i < len(response.Hits.Hits); i++ {
			sr.Items = append(sr.Items, t.buildContentItem(response.Hits.Hits[i], highlight))
		}
		for j := 0; j < len(response.Aggregations.DocCounts.Buckets); j++ {
			sr.ContentTypes = append(sr.ContentTypes, buildContentTypes(response.Aggregations.DocCounts.Buckets[j]))
		}
		for k := 0; k < len(response.Suggest.SearchSuggest); k++ {
			for _, option := range response.Suggest.SearchSuggest[k].Options {
				sr.Suggestions = append(sr.Suggestions, option.Text)
			}
		}
		took += response.Took
	}
	sr.Took = took
	return sr
}

func (t *LegacyTransformer) buildContentItem(doc models.ESResponseHitLegacy, highlight bool) models.ContentItemLegacy {
	ci := models.ContentItemLegacy{
		Description: t.buildDescription(doc, highlight),
		Type:        doc.Source.Type,
		URI:         doc.Source.URI,
	}

	return ci
}

func (t *LegacyTransformer) buildDescription(doc models.ESResponseHitLegacy, highlight bool) models.DescriptionLegacy {
	sd := doc.Source.Description
	hl := doc.Highlight

	des := models.DescriptionLegacy{
		Summary:           sd.Summary,
		NextRelease:       sd.NextRelease,
		Unit:              sd.Unit,
		PreUnit:           sd.PreUnit,
		Keywords:          sd.Keywords,
		ReleaseDate:       sd.ReleaseDate,
		Edition:           sd.Edition,
		LatestRelease:     sd.LatestRelease,
		Language:          sd.Language,
		Contact:           sd.Contact,
		DatasetID:         sd.DatasetID,
		Source:            sd.Source,
		Title:             sd.Title,
		MetaDescription:   sd.MetaDescription,
		NationalStatistic: sd.NationalStatistic,
		Headline1:         sd.Headline1,
		Headline2:         sd.Headline2,
		Headline3:         sd.Headline3,
	}

	if highlight {
		des.Highlight = &models.HighlightObjLegacy{
			DatasetID:       t.overlaySingleItem(hl.DescriptionDatasetID, sd.DatasetID, highlight),
			Edition:         t.overlaySingleItem(hl.DescriptionEdition, sd.Edition, highlight),
			Keywords:        t.overlayItemList(hl.DescriptionKeywords, sd.Keywords, highlight),
			MetaDescription: t.overlaySingleItem(hl.DescriptionMeta, sd.MetaDescription, highlight),
			Summary:         t.overlaySingleItem(hl.DescriptionSummary, sd.Summary, highlight),
			Title:           t.overlaySingleItem(hl.DescriptionTitle, sd.Title, highlight),
		}
	}

	return des
}

func (t *LegacyTransformer) overlaySingleItem(hl []*string, def string, highlight bool) (overlaid string) {
	if highlight && hl != nil && len(hl) > 0 {
		overlaid = *(hl)[0]
	}
	return
}

func (t *LegacyTransformer) overlayItemList(hlList, defaultList []*string, highlight bool) []*string {
	if defaultList == nil || hlList == nil {
		return nil
	}

	overlaid := make([]*string, len(defaultList))
	copy(overlaid, defaultList)
	if highlight {
		for _, hl := range hlList {
			unformatted := t.higlightReplacer.Replace(*hl)
			for i, defItem := range overlaid {
				if *defItem == unformatted {
					overlaid[i] = hl
				}
			}
		}
	}

	return overlaid
}

func buildContentTypes(bucket models.ESBucketLegacy) models.ContentType {
	return models.ContentType{
		Type:  bucket.Key,
		Count: bucket.Count,
	}
}

func buildAdditionalSuggestionList(query string) []string {
	regex := regexp.MustCompile(`"[^"]*"|\S+`)

	queryTerms := []string{}
	for _, match := range regex.FindAllStringSubmatch(query, -1) {
		queryTerms = append(queryTerms, match[0])
	}
	return queryTerms
}

func numberOfSearchTerms(query string) int {
	st := strings.Fields(query)
	return len(st)
}

// TransformSearchResponse transforms an elastic search 7.x response
func (t *Transformer) TransformSearchResponse(
	ctx context.Context, responseData []byte,
	query string, highlight bool) ([]byte, error) {
	var esResponse models.EsResponses

	err := json.Unmarshal(responseData, &esResponse)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode elastic search 7x response")
	}

	if len(esResponse.Responses) < 1 {
		return nil, errors.New("Response to be 7x transformed contained 0 items")
	}

	sr := t.transform(&esResponse, highlight)

	needAdditionalSuggestions := numberOfSearchTerms(query)
	if needAdditionalSuggestions > 1 {
		as := buildAdditionalSuggestionList(query)
		sr.AdditionSuggestions = as
	}

	transformedData, err := json.Marshal(sr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode transformed response")
	}
	return transformedData, nil
}

// Transform the raw ES to search response
func (t *Transformer) transform(esresponses *models.EsResponses, highlight bool) models.SearchResponse {
	search7xResponse := models.SearchResponse{
		Items:        []models.ESSourceDocument{},
		ContentTypes: []models.ContentType{},
	}
	var took int
	for _, response := range esresponses.Responses {
		for i := 0; i < len(response.Hits.Hits); i++ {
			search7xResponse.Items = append(search7xResponse.Items, t.buildContentItem(response.Hits.Hits[i], highlight))
		}
		for j := 0; j < len(response.Aggregations.DocCounts.Buckets); j++ {
			search7xResponse.ContentTypes = append(search7xResponse.ContentTypes, models.ContentType{
				Type:  response.Aggregations.DocCounts.Buckets[j].Key,
				Count: response.Aggregations.DocCounts.Buckets[j].Count,
			})
		}
		for _, suggestion := range response.Suggest.SearchSuggest {
			search7xResponse.Suggestions = append(search7xResponse.Suggestions, suggestion.Text)
			for _, option := range suggestion.Options {
				search7xResponse.Suggestions = append(search7xResponse.Suggestions, option.Text)
			}
		}
		took += response.Took
	}
	search7xResponse.Took = took
	return search7xResponse
}

func (t *Transformer) buildContentItem(doc models.ESResponseHit, highlight bool) models.ESSourceDocument {
	hl := doc.Highlight
	esDoc := models.ESSourceDocument{
		CDID:            doc.Source.CDID,
		DataType:        doc.Source.DataType,
		DatasetID:       doc.Source.DatasetID,
		URI:             doc.Source.URI,
		Keywords:        doc.Source.Keywords,
		MetaDescription: doc.Source.MetaDescription,
		ReleaseDate:     doc.Source.ReleaseDate,
		Summary:         doc.Source.Summary,
		Title:           doc.Source.Title,
		Topics:          doc.Source.Topics,
		Cancelled:       doc.Source.Cancelled,
		Finalised:       doc.Source.Finalised,
		ProvisionalDate: doc.Source.ProvisionalDate,
		Published:       doc.Source.Published,
		DateChanges:     doc.Source.DateChanges,
	}

	if highlight {
		esDoc.Highlight = &models.HighlightObj{
			DatasetID:       t.overlaySingleItem(hl.DescriptionDatasetID, doc.Source.DatasetID, highlight),
			Keywords:        t.overlayItemList(hl.DescriptionKeywords, doc.Source.Keywords, highlight),
			MetaDescription: t.overlaySingleItem(hl.DescriptionMeta, doc.Source.MetaDescription, highlight),
			Summary:         t.overlaySingleItem(hl.DescriptionSummary, doc.Source.Summary, highlight),
			Title:           t.overlaySingleItem(hl.DescriptionTitle, doc.Source.Title, highlight),
		}
	}

	return esDoc
}

func (t *Transformer) overlaySingleItem(hl []*string, def string, highlight bool) (overlaid string) {
	if highlight && hl != nil && len(hl) > 0 {
		overlaid = *(hl)[0]
	}
	return
}

func (t *Transformer) overlayItemList(hlList []*string, defaultList []string, highlight bool) []*string {
	if defaultList == nil || hlList == nil {
		return nil
	}
	var defaultListptr []*string
	for i := 0; i < len(defaultList); i++ {
		defaultListptr = append(defaultListptr, &defaultList[i])
	}

	overlaid := make([]*string, len(defaultListptr))
	copy(overlaid, defaultListptr)
	if highlight {
		for _, hl := range hlList {
			unformatted := t.higlightReplacer.Replace(*hl)
			for i, defItem := range overlaid {
				if *defItem == unformatted {
					overlaid[i] = hl
				}
			}
		}
	}

	return overlaid
}
