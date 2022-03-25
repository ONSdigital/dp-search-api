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
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
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
	URI         string             `json:"uri"`
	Description ReleaseDescription `json:"description"`
	Highlight   *highlight         `json:"highlight,omitempty"`
}

type ReleaseDescription struct {
	Title              string          `json:"title"`
	Summary            string          `json:"summary"`
	ReleaseDate        string          `json:"release_date"`
	Published          bool            `json:"published"`
	Cancelled          bool            `json:"cancelled"`
	Finalised          bool            `json:"finalised"`
	Postponed          bool            `json:"postponed"`
	Census             bool            `json:"census"`
	NationalStatistic  bool            `json:"national_statistic"`
	Keywords           []string        `json:"keywords,omitempty"`
	NextRelease        string          `json:"next_release,omitempty"`
	ProvisionalDate    string          `json:"provisional_date,omitempty"`
	CancellationNotice []string        `json:"cancellation_notice,omitempty"`
	Edition            string          `json:"edition,omitempty"`
	DatasetID          string          `json:"dataset_id,omitempty"`
	LatestRelease      *bool           `json:"latest_release,omitempty"`
	MetaDescription    string          `json:"meta_description,omitempty"`
	Language           string          `json:"language,omitempty"`
	Source             string          `json:"source,omitempty"`
	Contact            *releaseContact `json:"contact,omitempty"`
}

type releaseContact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone,omitempty"`
	Email     string `json:"email"`
}

type highlight struct {
	DatasetID       string   `json:"dataset_id,omitempty"`
	Edition         string   `json:"edition,omitempty"`
	Keywords        []string `json:"keywords,omitempty"`
	MetaDescription string   `json:"meta_description,omitempty"`
	Summary         string   `json:"summary,omitempty"`
	Title           string   `json:"title,omitempty"`
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
		Title              string          `json:"title"`
		Summary            string          `json:"summary"`
		ReleaseDate        string          `json:"releaseDate,omitempty"`
		Published          bool            `json:"published"`
		Cancelled          bool            `json:"cancelled"`
		Finalised          bool            `json:"finalised"`
		Topics             []string        `json:"topics"`
		NationalStatistic  bool            `json:"nationalStatistic,omitempty"`
		Keywords           []string        `json:"keywords,omitempty"`
		NextRelease        string          `json:"nextRelease,omitempty"`
		CancellationNotice []string        `json:"cancellationNotice"`
		ProvisionalDate    string          `json:"provisionalDate"`
		Edition            string          `json:"edition,omitempty"`
		DatasetID          string          `json:"datasetId,omitempty"`
		LatestRelease      bool            `json:"latestRelease"`
		MetaDescription    string          `json:"metaDescription,omitempty"`
		Language           string          `json:"language,omitempty"`
		Source             string          `json:"source,omitempty"`
		Contact            *releaseContact `json:"contact,omitempty"`
	} `json:"description"`
}

type dateChange struct {
	PreviousDate string `json:"previousDate,omitempty"`
	ChangeNotice string `json:"changeNotice,omitempty"`
}

type ESReleaseHighlight struct {
	DescriptionTitle     []string `json:"description.title"`
	DescriptionEdition   []string `json:"description.edition"`
	DescriptionSummary   []string `json:"description.summary"`
	DescriptionMeta      []string `json:"description.metaDescription"`
	DescriptionKeywords  []string `json:"description.keywords"`
	DescriptionDatasetID []string `json:"description.datasetId"`
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
		Limit:     10,
		Offset:    0,
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
			// The following 3 need to be added to source document (and indexed)
			Published:         sd.Published,
			Cancelled:         sd.Cancelled,
			Finalised:         sd.Finalised,
			Postponed:         isPostponed(hit.Source),
			Census:            isCensus(hit.Source),
			NationalStatistic: sd.NationalStatistic,
			Keywords:          sd.Keywords,
			NextRelease:       sd.NextRelease,
			// The following 3 need to be added to source document
			ProvisionalDate:    sd.ProvisionalDate,
			CancellationNotice: sd.CancellationNotice,
			Edition:            sd.Edition,
			DatasetID:          sd.DatasetID,
			// The following 1 needs to be added to source document
			LatestRelease:   &sd.LatestRelease,
			MetaDescription: sd.MetaDescription,
			Language:        sd.Language,
			Contact:         sd.Contact,
			Source:          sd.Source,
		},
	}

	if highlightOn {
		r.Highlight = &highlight{
			DatasetID:       t.overlayItem(hl.DescriptionDatasetID, sd.DatasetID, highlightOn),
			Edition:         t.overlayItem(hl.DescriptionEdition, sd.Edition, highlightOn),
			Keywords:        t.overlayList(hl.DescriptionKeywords, sd.Keywords, highlightOn),
			MetaDescription: t.overlayItem(hl.DescriptionMeta, sd.MetaDescription, highlightOn),
			Summary:         t.overlayItem(hl.DescriptionSummary, sd.Summary, highlightOn),
			Title:           t.overlayItem(hl.DescriptionTitle, sd.Title, highlightOn),
		}
	}

	return r
}

func isCensus(_ ESReleaseSourceDocument) bool {
	return false
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
