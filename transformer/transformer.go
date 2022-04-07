package transformer

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Transformer represents an instance of the ResponseTransformer interface
type LegacyTransformer struct {
	higlightReplacer *strings.Replacer
}

// Structs representing the transformed response
type SearchResponse struct {
	Count               int           `json:"count"`
	Took                int           `json:"took"`
	ContentTypes        []ContentType `json:"content_types"`
	Items               []ContentItem `json:"items"`
	Suggestions         []string      `json:"suggestions,omitempty"`
	AdditionSuggestions []string      `json:"additional_suggestions,omitempty"`
}

type ContentType struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type ContentItem struct {
	Description description `json:"description"`
	Type        string      `json:"type"`
	URI         string      `json:"uri"`
}

type description struct {
	Contact           *contact      `json:"contact,omitempty"`
	DatasetID         string        `json:"dataset_id,omitempty"`
	Edition           string        `json:"edition,omitempty"`
	Headline1         string        `json:"headline1,omitempty"`
	Headline2         string        `json:"headline2,omitempty"`
	Headline3         string        `json:"headline3,omitempty"`
	Highlight         *highlightObj `json:"highlight,omitempty"`
	Keywords          *[]string     `json:"keywords,omitempty"`
	LatestRelease     *bool         `json:"latest_release,omitempty"`
	Language          string        `json:"language,omitempty"`
	MetaDescription   string        `json:"meta_description,omitempty"`
	NationalStatistic *bool         `json:"national_statistic,omitempty"`
	NextRelease       string        `json:"next_release,omitempty"`
	PreUnit           string        `json:"pre_unit,omitempty"`
	ReleaseDate       string        `json:"release_date,omitempty"`
	Source            string        `json:"source,omitempty"`
	Summary           string        `json:"summary"`
	Title             string        `json:"title"`
	Unit              string        `json:"unit,omitempty"`
}

type contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone,omitempty"`
	Email     string `json:"email"`
}

type highlightObj struct {
	DatasetID       string    `json:"dataset_id,omitempty"`
	Edition         string    `json:"edition,omitempty"`
	Keywords        *[]string `json:"keywords,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	Summary         string    `json:"summary,omitempty"`
	Title           string    `json:"title,omitempty"`
}

// Structs representing the raw elastic search response

type ESResponse struct {
	Responses []ESResponseItem `json:"responses"`
}

type ESResponseItem struct {
	Took         int                    `json:"took"`
	Hits         ESResponseHits         `json:"hits"`
	Aggregations ESResponseAggregations `json:"aggregations"`
	Suggest      ESSuggest              `json:"suggest"`
}

type ESResponseHits struct {
	Total int
	Hits  []ESResponseHit `json:"hits"`
}

type ESResponseHit struct {
	Source    ESSourceDocument `json:"_source"`
	Highlight ESHighlight      `json:"highlight"`
}

type ESSourceDocument struct {
	Description struct {
		Summary           string    `json:"summary"`
		NextRelease       string    `json:"nextRelease,omitempty"`
		Unit              string    `json:"unit,omitempty"`
		Keywords          *[]string `json:"keywords,omitempty"`
		ReleaseDate       string    `json:"releaseDate,omitempty"`
		Edition           string    `json:"edition,omitempty"`
		LatestRelease     *bool     `json:"latestRelease,omitempty"`
		Language          string    `json:"language,omitempty"`
		Contact           *contact  `json:"contact,omitempty"`
		DatasetID         string    `json:"datasetId,omitempty"`
		Source            string    `json:"source,omitempty"`
		Title             string    `json:"title"`
		MetaDescription   string    `json:"metaDescription,omitempty"`
		NationalStatistic *bool     `json:"nationalStatistic,omitempty"`
		PreUnit           string    `json:"preUnit,omitempty"`
		Headline1         string    `json:"headline1,omitempty"`
		Headline2         string    `json:"headline2,omitempty"`
		Headline3         string    `json:"headline3,omitempty"`
	} `json:"description"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type ESHighlight struct {
	DescriptionTitle     *[]string `json:"description.title"`
	DescriptionEdition   *[]string `json:"description.edition"`
	DescriptionSummary   *[]string `json:"description.summary"`
	DescriptionMeta      *[]string `json:"description.metaDescription"`
	DescriptionKeywords  *[]string `json:"description.keywords"`
	DescriptionDatasetID *[]string `json:"description.datasetId"`
}

type ESResponseAggregations struct {
	DocCounts struct {
		Buckets []ESBucket `json:"buckets"`
	} `json:"docCounts"`
}

type ESBucket struct {
	Key   string `json:"key"`
	Count int    `json:"doc_count"`
}

type ESSuggest struct {
	SearchSuggest []ESSearchSuggest `json:"search_suggest"`
}

type ESSearchSuggest struct {
	Options []ESSearchSuggestOptions `json:"options"`
}

type ESSearchSuggestOptions struct {
	Text string `json:"text"`
}

// New returns a new instance of Transformer
func New() *LegacyTransformer {
	highlightReplacer := strings.NewReplacer("<em class=\"highlight\">", "", "</em>", "")
	return &LegacyTransformer{
		higlightReplacer: highlightReplacer,
	}
}

// TransformSearchResponse transforms an elastic search response into a structure that matches the v1 api specification
func (t *LegacyTransformer) TransformSearchResponse(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error) {
	var source ESResponse

	err := json.Unmarshal(responseData, &source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode elastic search response")
	}

	if len(source.Responses) < 1 {
		return nil, errors.New("Response to be transformed contained 0 items")
	}

	sr := t.transform(&source, highlight)

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

func (t *LegacyTransformer) transform(source *ESResponse, highlight bool) SearchResponse {
	sr := SearchResponse{
		Count:        source.Responses[0].Hits.Total,
		Items:        []ContentItem{},
		ContentTypes: []ContentType{},
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

func (t *LegacyTransformer) buildContentItem(doc ESResponseHit, highlight bool) ContentItem {
	ci := ContentItem{
		Description: t.buildDescription(doc, highlight),
		Type:        doc.Source.Type,
		URI:         doc.Source.URI,
	}

	return ci
}

func (t *LegacyTransformer) buildDescription(doc ESResponseHit, highlight bool) description {
	sd := doc.Source.Description
	hl := doc.Highlight

	des := description{
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
		des.Highlight = &highlightObj{
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

func buildContentTypes(bucket ESBucket) ContentType {
	return ContentType{
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
