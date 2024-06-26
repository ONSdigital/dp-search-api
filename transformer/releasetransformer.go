package transformer

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-search-api/api"
	"github.com/ONSdigital/dp-search-api/query"
)

type ReleaseTransformer struct {
	higlightReplacer *strings.Replacer
}

type SearchReleaseResponse struct {
	Took      int       `json:"took"`
	Breakdown Breakdown `json:"breakdown"`
	Releases  []Release `json:"releases"`
}

type Breakdown struct {
	Total       int `json:"total"`
	Provisional int `json:"provisional,omitempty"`
	Confirmed   int `json:"confirmed,omitempty"`
	Postponed   int `json:"postponed,omitempty"`
	Published   int `json:"published,omitempty"`
	Cancelled   int `json:"cancelled,omitempty"`
	Census      int `json:"census,omitempty"`
}

type Release struct {
	URI         string              `json:"uri"`
	DateChanges []ReleaseDateChange `json:"date_changes"`
	Description ReleaseDescription  `json:"description"`
	Highlight   *highlight          `json:"highlight,omitempty"`
}

type ReleaseDateChange struct {
	ChangeNotice string `json:"change_notice"`
	Date         string `json:"previous_date"`
}

type ReleaseDescription struct {
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	ReleaseDate     string   `json:"release_date"`
	Published       bool     `json:"published"`
	Cancelled       bool     `json:"cancelled"`
	Finalised       bool     `json:"finalised"`
	Postponed       bool     `json:"postponed"`
	Census          bool     `json:"census"`
	Keywords        []string `json:"keywords,omitempty"`
	ProvisionalDate string   `json:"provisional_date,omitempty"`
	Language        string   `json:"language,omitempty"`
	CanonicalTopic  string   `json:"canonical_topic,omitempty"`
}

type highlight struct {
	Keywords []string `json:"keywords,omitempty"`
	Summary  string   `json:"summary,omitempty"`
	Title    string   `json:"title,omitempty"`
}

// Structs representing the raw elastic search response for ES7.10

type ESReleaseResponse struct {
	Responses []ESReleaseResponseItem `json:"responses"`
}

type ESReleaseResponseItem struct {
	Took         int                           `json:"took"`
	TimedOut     bool                          `json:"timed_out"`
	Hits         ESReleaseResponseSummary      `json:"hits"`
	Aggregations ESReleaseResponseAggregations `json:"aggregations"`
}

type ESReleaseResponseSummary struct {
	Total struct {
		Value int `json:"value"`
	} `json:"total"`
	Hits []ESReleaseResponseHit `json:"hits"`
}

type ESReleaseResponseHit struct {
	Source    ESReleaseSourceDocument `json:"_source"`
	Highlight ESReleaseHighlight      `json:"highlight"`
}

type ESReleaseSourceDocument struct {
	URI             string       `json:"uri"`
	Title           string       `json:"title"`
	Summary         string       `json:"summary"`
	ReleaseDate     string       `json:"release_date,omitempty"`
	Published       bool         `json:"published"`
	Cancelled       bool         `json:"cancelled"`
	Finalised       bool         `json:"finalised"`
	Survey          string       `json:"survey"`
	Keywords        []string     `json:"keywords,omitempty"`
	Language        string       `json:"language,omitempty"`
	DateChanges     []dateChange `json:"date_changes,omitempty"`
	ProvisionalDate string       `json:"provisional_date,omitempty"`
}

type dateChange struct {
	PreviousDate string `json:"previous_date,omitempty"`
	ChangeNotice string `json:"change_notice,omitempty"`
}

type ESReleaseHighlight struct {
	Title    []string `json:"title"`
	Summary  []string `json:"summary"`
	Keywords []string `json:"keywords"`
}

type aggName string
type ESReleaseResponseAggregations map[aggName]aggregation

type bucketName string
type aggregation struct {
	Buckets map[bucketName]bucketContents `json:"buckets"`
}

type bucketContents struct {
	Count     int         `json:"doc_count"`
	Breakdown aggregation `json:"breakdown"`
}

func NewReleaseTransformer() api.ReleaseResponseTransformer {
	highlightReplacer := strings.NewReplacer("<em class=\"highlight\">", "", "</em>", "")
	return &ReleaseTransformer{
		higlightReplacer: highlightReplacer,
	}
}

const numberOfReleaseQueries = 2

// TransformSearchResponse transforms an elastic search response to a release query into a serialised ReleaseResponse
func (t *ReleaseTransformer) TransformSearchResponse(_ context.Context, responseData []byte, req query.ReleaseSearchRequest, highlight bool) ([]byte, error) {
	var (
		source      ESReleaseResponse
		highlighter *strings.Replacer
	)

	err := json.Unmarshal(responseData, &source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode elastic search response")
	}

	if len(source.Responses) != numberOfReleaseQueries {
		return nil, errors.New("invalid number of responses from ElasticSearch query")
	}

	sr := SearchReleaseResponse{
		Took:      source.Responses[0].Took + source.Responses[1].Took,
		Breakdown: breakdown(source, req),
	}

	if highlight {
		highlighter = t.higlightReplacer
	}
	for i := 0; i < len(source.Responses[0].Hits.Hits); i++ {
		sr.Releases = append(sr.Releases, buildRelease(source.Responses[0].Hits.Hits[i], highlighter))
	}

	transformedData, err := json.Marshal(sr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode transformed response")
	}

	return transformedData, nil
}

func breakdown(source ESReleaseResponse, req query.ReleaseSearchRequest) Breakdown {
	b := Breakdown{Total: source.Responses[0].Hits.Total.Value}

	switch req.Type {
	case query.Upcoming:
		b.Provisional = source.Responses[0].Aggregations["breakdown"].Buckets["provisional"].Count
		b.Confirmed = source.Responses[0].Aggregations["breakdown"].Buckets["confirmed"].Count
		b.Postponed = source.Responses[0].Aggregations["breakdown"].Buckets["postponed"].Count

		b.Published = source.Responses[1].Aggregations["release_types"].Buckets["published"].Count
		b.Cancelled = source.Responses[1].Aggregations["release_types"].Buckets["cancelled"].Count
	case query.Published:
		b.Published = source.Responses[0].Hits.Total.Value
		b.Cancelled = source.Responses[1].Aggregations["release_types"].Buckets["cancelled"].Count

		b.Provisional = source.Responses[1].Aggregations["release_types"].Buckets["upcoming"].Breakdown.Buckets["provisional"].Count
		b.Confirmed = source.Responses[1].Aggregations["release_types"].Buckets["upcoming"].Breakdown.Buckets["confirmed"].Count
		b.Postponed = source.Responses[1].Aggregations["release_types"].Buckets["upcoming"].Breakdown.Buckets["postponed"].Count
	case query.Cancelled:
		b.Cancelled = source.Responses[0].Hits.Total.Value
		b.Published = source.Responses[1].Aggregations["release_types"].Buckets["published"].Count

		b.Provisional = source.Responses[1].Aggregations["release_types"].Buckets["upcoming"].Breakdown.Buckets["provisional"].Count
		b.Confirmed = source.Responses[1].Aggregations["release_types"].Buckets["upcoming"].Breakdown.Buckets["confirmed"].Count
		b.Postponed = source.Responses[1].Aggregations["release_types"].Buckets["upcoming"].Breakdown.Buckets["postponed"].Count
	}

	b.Census = source.Responses[0].Aggregations["census"].Buckets["census"].Count

	return b
}

func buildRelease(hit ESReleaseResponseHit, highlighter *strings.Replacer) Release {
	sd := hit.Source
	hl := hit.Highlight

	r := Release{
		URI: hit.Source.URI,
		Description: ReleaseDescription{
			Title:           sd.Title,
			Summary:         sd.Summary,
			ReleaseDate:     sd.ReleaseDate,
			Published:       sd.Published,
			Cancelled:       sd.Cancelled,
			Finalised:       sd.Finalised,
			Postponed:       isPostponed(sd),
			Census:          isCensus(sd),
			Keywords:        sd.Keywords,
			Language:        sd.Language,
			ProvisionalDate: sd.ProvisionalDate,
		},
	}

	for _, dc := range hit.Source.DateChanges {
		r.DateChanges = append(r.DateChanges, ReleaseDateChange{Date: dc.PreviousDate, ChangeNotice: dc.ChangeNotice})
	}

	if highlighter != nil {
		r.Highlight = &highlight{
			Keywords: overlayList(hl.Keywords, sd.Keywords, highlighter),
			Summary:  overlayItem(hl.Summary, sd.Summary, highlighter),
			Title:    overlayItem(hl.Title, sd.Title, highlighter),
		}
	}

	return r
}

func isPostponed(release ESReleaseSourceDocument) bool {
	return release.Finalised && len(release.DateChanges) > 0
}

func isCensus(release ESReleaseSourceDocument) bool {
	return release.Survey == "census"
}

func overlayItem(hl []string, def string, highlighter *strings.Replacer) string {
	if highlighter != nil && len(hl) > 0 {
		return hl[0]
	}

	return def
}

func overlayList(hlList, defaultList []string, highlighter *strings.Replacer) []string {
	if defaultList == nil || hlList == nil {
		return nil
	}

	overlaid := make([]string, len(defaultList))
	copy(overlaid, defaultList)
	if highlighter != nil {
		for _, hl := range hlList {
			unformatted := highlighter.Replace(hl)
			for i, defItem := range overlaid {
				if defItem == unformatted {
					overlaid[i] = hl
				}
			}
		}
	}

	return overlaid
}
