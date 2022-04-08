package transformer

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/ONSdigital/dp-search-api/models"
	"github.com/pkg/errors"
)

// LegacyTransformer represents an instance of the ResponseTransformer interface
type LegacyTransformer struct {
	higlightReplacer *strings.Replacer
}

// NewLegacy returns a new instance of Transformer
func NewLegacy() *LegacyTransformer {
	highlightReplacer := strings.NewReplacer("<em class=\"highlight\">", "", "</em>", "")
	return &LegacyTransformer{
		higlightReplacer: highlightReplacer,
	}
}

// Transformer represents an instance of the ResponseTransformer interface for ES7x
type Transformer struct {
	higlightReplacer *strings.Replacer
}

// New7x returns a new instance of Transformer7x
func New() *Transformer {
	highlightReplacer := strings.NewReplacer("<em class=\"highlight\">", "", "</em>", "")
	return &Transformer{
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
		ContentTypes: []models.ContentTypeLegacy{},
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

func (t *LegacyTransformer) overlaySingleItem(hl *[]string, def string, highlight bool) (overlaid string) {
	if highlight && hl != nil && len(*hl) > 0 {
		overlaid = (*hl)[0]
	}
	return
}

func (t *LegacyTransformer) overlayItemList(hlList, defaultList *[]string, highlight bool) *[]string {
	if defaultList == nil || hlList == nil {
		return nil
	}

	overlaid := make([]string, len(*defaultList))
	copy(overlaid, *defaultList)
	if highlight {
		for _, hl := range *hlList {
			unformatted := t.higlightReplacer.Replace(hl)
			for i, defItem := range overlaid {
				if defItem == unformatted {
					overlaid[i] = hl
				}
			}
		}
	}

	return &overlaid
}

func buildContentTypes(bucket models.ESBucketLegacy) models.ContentTypeLegacy {
	return models.ContentTypeLegacy{
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
	var esResponse models.Es7xResponse

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
func (t *Transformer) transform(es7xresponse *models.Es7xResponse, highlight bool) models.Search7xResponse {
	var search7xResponse = models.Search7xResponse{
		Took:        es7xresponse.Responses[0].Took,
		Items:       es7xresponse.Responses[0].Hits.Hits[0].Source,
		Suggestions: es7xresponse.Responses[0].Suggest,
	}
	return search7xResponse
}
