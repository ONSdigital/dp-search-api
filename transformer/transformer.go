package transformer

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

// Transformer represents an instance of the ResponseTransformer interface
type Transformer struct {
	higlightReplacer strings.Replacer
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
	Contact           *contact  `json:"contact,omitempty"`
	DatasetID         string    `json:"dataset_id,omitempty"`
	Edition           string    `json:"edition,omitempty"`
	Headline1         string    `json:"headline1,omitempty"`
	Headline2         string    `json:"headline2,omitempty"`
	Headline3         string    `json:"headline3,omitempty"`
	Keywords          *[]string `json:"keywords,omitempty"`
	LatestRelease     *bool     `json:"latest_release,omitempty"`
	Language          string    `json:"language,omitempty"`
	MetaDescription   string    `json:"meta_description,omitempty"`
	NationalStatistic *bool     `json:"national_statistic,omitempty"`
	NextRelease       string    `json:"next_release,omitempty"`
	PreUnit           string    `json:"pre_unit,omitempty"`
	ReleaseDate       string    `json:"release_date,omitempty"`
	Source            string    `json:"source,omitempty"`
	Summary           string    `json:"summary"`
	Title             string    `json:"title"`
	Unit              string    `json:"unit,omitempty"`
}

type contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone,omitempty"`
	Email     string `json:"email"`
}

// Structs representing the raw elastic search response

type ESResponse struct {
	Responses []esResponseItem `json:"responses"`
}

type esResponseItem struct {
	Took         int                    `json:"took"`
	Hits         esResponseHits         `json:"hits"`
	Aggregations esResponseAggregations `json:"aggregations"`
	Suggest      esSuggest              `json:"suggest"`
}

type esResponseHits struct {
	Total int
	Hits  []esResponseHit `json:"hits"`
}

type esResponseHit struct {
	Source    esSourceDocument `json:"_source"`
	Highlight esHighlight      `json:"highlight"`
}

type esSourceDocument struct {
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

type esHighlight struct {
	DescriptionTitle     *[]string `json:"description.title"`
	DescriptionEdition   *[]string `json:"description.edition"`
	DescriptionSummary   *[]string `json:"description.summary"`
	DescriptionMeta      *[]string `json:"description.metaDescription"`
	DescriptionKeywords  *[]string `json:"description.keywords"`
	DescriptionDatasetID *[]string `json:"description.datasetId"`
}

type esResponseAggregations struct {
	DocCounts struct {
		Buckets []esBucket `json:"buckets"`
	} `json:"docCounts"`
}

type esBucket struct {
	Key   string `json:"key"`
	Count int    `json:"doc_count"`
}

type esSuggest struct {
	SearchSuggest []esSearchSuggest `json:"search_suggest"`
}

type esSearchSuggest struct {
	Options []esSearchSuggestOptions `json:"options"`
}

type esSearchSuggestOptions struct {
	Text string `json:"text"`
}

// New returns a new instance of Transformer
func New() *Transformer {
	highlightReplacer := strings.NewReplacer("<em class=\"highlight\">", "", "</em>", "")
	return &Transformer{
		higlightReplacer: *highlightReplacer,
	}
}

// TransformSearchResponse transforms an elastic search response into a structure that matches the v1 api specification
func (t *Transformer) TransformSearchResponse(ctx context.Context, responseData []byte, query string, highlight bool) ([]byte, error) {
	var source esResponse

	err := json.Unmarshal(responseData, &source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode elastic search response")
	}

	if len(source.Responses) < 1 {
		return nil, errors.New("Response to be transformed contained 0 items")
	}

	sr := t.transform(&source, highlight)

	if sr.Count == 0 {
		as := buildAdditionalSuggestionList(query)
		sr.AdditionSuggestions = as
	}

	transformedData, err := json.Marshal(sr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode transformed response")
	}
	return transformedData, nil
}

func (t *Transformer) transform(source *esResponse, highlight bool) SearchResponse {
	sr := SearchResponse{
		Count:        source.Responses[0].Hits.Total,
		Items:        []ContentItem{},
		ContentTypes: []ContentType{},
	}
	var took int = 0
	for _, response := range source.Responses {
		for _, doc := range response.Hits.Hits {
			sr.Items = append(sr.Items, t.buildContentItem(doc, highlight))
		}
		for _, bucket := range response.Aggregations.DocCounts.Buckets {
			sr.ContentTypes = append(sr.ContentTypes, buildContentTypes(bucket))
		}
		for _, suggest := range response.Suggest.SearchSuggest {
			for _, option := range suggest.Options {
				sr.Suggestions = append(sr.Suggestions, option.Text)
			}
		}
		took += response.Took
	}
	sr.Took = took
	return sr
}

func (t *Transformer) buildContentItem(doc esResponseHit, highlight bool) ContentItem {
	ci := ContentItem{
		Description: t.buildDescription(doc, highlight),
		Type:        doc.Source.Type,
		URI:         doc.Source.URI,
	}

	return ci
}

func (t *Transformer) buildDescription(doc esResponseHit, highlight bool) description {
	sd := doc.Source.Description
	hl := doc.Highlight

	return description{
		Summary:           t.overlaySingleItem(hl.DescriptionSummary, sd.Summary, highlight),
		NextRelease:       sd.NextRelease,
		Unit:              sd.Unit,
		PreUnit:           sd.PreUnit,
		Keywords:          t.overlayItemList(hl.DescriptionKeywords, sd.Keywords, highlight),
		ReleaseDate:       sd.ReleaseDate,
		Edition:           t.overlaySingleItem(hl.DescriptionEdition, sd.Edition, highlight),
		LatestRelease:     sd.LatestRelease,
		Language:          sd.Language,
		Contact:           sd.Contact,
		DatasetID:         t.overlaySingleItem(hl.DescriptionDatasetID, sd.DatasetID, highlight),
		Source:            sd.Source,
		Title:             t.overlaySingleItem(hl.DescriptionTitle, sd.Title, highlight),
		MetaDescription:   t.overlaySingleItem(hl.DescriptionMeta, sd.MetaDescription, highlight),
		NationalStatistic: sd.NationalStatistic,
		Headline1:         sd.Headline1,
		Headline2:         sd.Headline2,
		Headline3:         sd.Headline3,
	}
}

func (t *Transformer) overlaySingleItem(hl *[]string, def string, highlight bool) string {
	overlaid := def
	if highlight && hl != nil && len(*hl) > 0 {
		overlaid = (*hl)[0]
	}
	return overlaid
}

func (t *Transformer) overlayItemList(hlList *[]string, defaultList *[]string, highlight bool) *[]string {
	if defaultList == nil {
		return nil
	}
	overlaid := *defaultList
	if highlight && hlList != nil {
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

func buildContentTypes(bucket esBucket) ContentType {
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
