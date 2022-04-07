package transformer

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
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
}

type highlight struct {
	Keywords []string `json:"keywords,omitempty"`
	Summary  string   `json:"summary,omitempty"`
	Title    string   `json:"title,omitempty"`
}

// Structs representing the raw elastic search response

type ESReleaseResponse struct {
	Took     int                      `json:"took"`
	TimedOut bool                     `json:"timed_out"`
	Hits     ESReleaseResponseSummary `json:"hits"`
}

type ESReleaseResponseSummary struct {
	Total int                    `json:"total"`
	Hits  []ESReleaseResponseHit `json:"hits"`
}

type ESReleaseResponseHit struct {
	Source    ESReleaseSourceDocument `json:"_source"`
	Highlight ESReleaseHighlight      `json:"highlight"`
}

type ESReleaseSourceDocument struct {
	URI         string       `json:"uri"`
	DateChanges []dateChange `json:"dateChanges,omitempty"`

	Description struct {
		Title       string   `json:"title"`
		Summary     string   `json:"summary"`
		ReleaseDate string   `json:"releaseDate,omitempty"`
		Published   bool     `json:"published"`
		Cancelled   bool     `json:"cancelled"`
		Finalised   bool     `json:"finalised"`
		Census      bool     `json:"census"`
		Keywords    []string `json:"keywords,omitempty"`
		Language    string   `json:"language,omitempty"`
	} `json:"description"`
}

type dateChange struct {
	PreviousDate string `json:"previousDate,omitempty"`
	ChangeNotice string `json:"changeNotice,omitempty"`
}

type ESReleaseHighlight struct {
	DescriptionTitle    []string `json:"description.title"`
	DescriptionSummary  []string `json:"description.summary"`
	DescriptionKeywords []string `json:"description.keywords"`
}

func NewReleaseTransformer() *ReleaseTransformer {
	highlightReplacer := strings.NewReplacer("<em class=\"highlight\">", "", "</em>", "")
	return &ReleaseTransformer{
		higlightReplacer: highlightReplacer,
	}
}

// TransformSearchResponse transforms an elastic search response into a structure that matches the v1 api specification
func (t *ReleaseTransformer) TransformSearchResponse(_ context.Context, responseData []byte, _ string, highlight bool) ([]byte, error) {
	var source ESReleaseResponse

	err := json.Unmarshal(responseData, &source)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode elastic search response")
	}

	sr := t.transform(&source, highlight)

	transformedData, err := json.Marshal(sr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode transformed response")
	}

	return transformedData, nil
}

func (t *ReleaseTransformer) transform(source *ESReleaseResponse, highlight bool) SearchReleaseResponse {
	sr := SearchReleaseResponse{
		Took:      source.Took,
		Breakdown: Breakdown{Total: source.Hits.Total},
		Releases:  []Release{},
	}

	for i := range source.Hits.Hits {
		sr.Releases = append(sr.Releases, t.buildRelease(source.Hits.Hits[i], highlight))
	}

	return sr
}

func (t *ReleaseTransformer) buildRelease(hit ESReleaseResponseHit, highlightOn bool) Release {
	sd := hit.Source.Description
	hl := hit.Highlight

	r := Release{
		URI: hit.Source.URI,
		Description: ReleaseDescription{
			Title:       sd.Title,
			Summary:     sd.Summary,
			ReleaseDate: sd.ReleaseDate,
			Published:   sd.Published,
			Cancelled:   sd.Cancelled,
			Finalised:   sd.Finalised,
			Postponed:   isPostponed(hit.Source),
			Census:      sd.Census,
			Keywords:    sd.Keywords,
			Language:    sd.Language,
		},
	}

	if highlightOn {
		r.Highlight = &highlight{
			Keywords: t.overlayList(hl.DescriptionKeywords, sd.Keywords, highlightOn),
			Summary:  t.overlayItem(hl.DescriptionSummary, sd.Summary, highlightOn),
			Title:    t.overlayItem(hl.DescriptionTitle, sd.Title, highlightOn),
		}
	}

	return r
}

func isPostponed(release ESReleaseSourceDocument) bool {
	return release.Description.Finalised && len(release.DateChanges) > 0
}

func (t *ReleaseTransformer) overlayItem(hl []string, def string, highlight bool) string {
	if highlight && len(hl) > 0 {
		return hl[0]
	}

	return def
}

func (t *ReleaseTransformer) overlayList(hlList, defaultList []string, highlight bool) []string {
	if defaultList == nil || hlList == nil {
		return nil
	}

	overlaid := make([]string, len(defaultList))
	copy(overlaid, defaultList)
	if highlight {
		for _, hl := range hlList {
			unformatted := t.higlightReplacer.Replace(hl)
			for i, defItem := range overlaid {
				if defItem == unformatted {
					overlaid[i] = hl
				}
			}
		}
	}

	return overlaid
}
