package models

// Structs representing the transformed response
type SearchResponseLegacy struct {
	Count               int                 `json:"count"`
	Took                int                 `json:"took"`
	ContentTypes        []ContentTypeLegacy `json:"content_types"`
	Items               []ContentItemLegacy `json:"items"`
	Suggestions         []string            `json:"suggestions,omitempty"`
	AdditionSuggestions []string            `json:"additional_suggestions,omitempty"`
}

type ContentTypeLegacy struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type ContentItemLegacy struct {
	Description DescriptionLegacy `json:"description"`
	Type        string            `json:"type"`
	URI         string            `json:"uri"`
}

type DescriptionLegacy struct {
	Contact           *contactLegacy      `json:"contact,omitempty"`
	DatasetID         string              `json:"dataset_id,omitempty"`
	Edition           string              `json:"edition,omitempty"`
	Headline1         string              `json:"headline1,omitempty"`
	Headline2         string              `json:"headline2,omitempty"`
	Headline3         string              `json:"headline3,omitempty"`
	Highlight         *HighlightObjLegacy `json:"highlight,omitempty"`
	Keywords          []*string           `json:"keywords,omitempty"`
	LatestRelease     *bool               `json:"latest_release,omitempty"`
	Language          string              `json:"language,omitempty"`
	MetaDescription   string              `json:"meta_description,omitempty"`
	NationalStatistic *bool               `json:"national_statistic,omitempty"`
	NextRelease       string              `json:"next_release,omitempty"`
	PreUnit           string              `json:"pre_unit,omitempty"`
	ReleaseDate       string              `json:"release_date,omitempty"`
	Source            string              `json:"source,omitempty"`
	Summary           string              `json:"summary"`
	Title             string              `json:"title"`
	Unit              string              `json:"unit,omitempty"`
}

type contactLegacy struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone,omitempty"`
	Email     string `json:"email"`
}

type HighlightObjLegacy struct {
	DatasetID       string    `json:"dataset_id,omitempty"`
	Edition         string    `json:"edition,omitempty"`
	Keywords        []*string `json:"keywords,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	Summary         string    `json:"summary,omitempty"`
	Title           string    `json:"title,omitempty"`
}

// Structs representing the raw elastic search response

type ESResponseLegacy struct {
	Responses []ESResponseItemLegacy `json:"responses"`
}

type ESResponseItemLegacy struct {
	Took         int                          `json:"took"`
	Hits         ESResponseHitsLegacy         `json:"hits"`
	Aggregations ESResponseAggregationsLegacy `json:"aggregations"`
	Suggest      ESSuggestLegacy              `json:"suggest"`
}

type ESResponseHitsLegacy struct {
	Total int
	Hits  []ESResponseHitLegacy `json:"hits"`
}

type ESResponseHitLegacy struct {
	Source    ESSourceDocumentLegacy `json:"_source"`
	Highlight ESHighlightLegacy      `json:"highlight"`
}

type ESSourceDocumentLegacy struct {
	Description struct {
		Summary           string         `json:"summary"`
		NextRelease       string         `json:"nextRelease,omitempty"`
		Unit              string         `json:"unit,omitempty"`
		Keywords          []*string      `json:"keywords,omitempty"`
		ReleaseDate       string         `json:"releaseDate,omitempty"`
		Edition           string         `json:"edition,omitempty"`
		LatestRelease     *bool          `json:"latestRelease,omitempty"`
		Language          string         `json:"language,omitempty"`
		Contact           *contactLegacy `json:"contact,omitempty"`
		DatasetID         string         `json:"datasetId,omitempty"`
		Source            string         `json:"source,omitempty"`
		Title             string         `json:"title"`
		MetaDescription   string         `json:"metaDescription,omitempty"`
		NationalStatistic *bool          `json:"nationalStatistic,omitempty"`
		PreUnit           string         `json:"preUnit,omitempty"`
		Headline1         string         `json:"headline1,omitempty"`
		Headline2         string         `json:"headline2,omitempty"`
		Headline3         string         `json:"headline3,omitempty"`
	} `json:"description"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type ESHighlightLegacy struct {
	DescriptionTitle     []*string `json:"description.title"`
	DescriptionEdition   []*string `json:"description.edition"`
	DescriptionSummary   []*string `json:"description.summary"`
	DescriptionMeta      []*string `json:"description.metaDescription"`
	DescriptionKeywords  []*string `json:"description.keywords"`
	DescriptionDatasetID []*string `json:"description.datasetId"`
}

type ESResponseAggregationsLegacy struct {
	DocCounts struct {
		Buckets []ESBucketLegacy `json:"buckets"`
	} `json:"docCounts"`
}

type ESBucketLegacy struct {
	Key   string `json:"key"`
	Count int    `json:"doc_count"`
}

type ESSuggestLegacy struct {
	SearchSuggest []ESSearchSuggestLegacy `json:"search_suggest"`
}

type ESSearchSuggestLegacy struct {
	Options []ESSearchSuggestOptionsLegacy `json:"options"`
}

type ESSearchSuggestOptionsLegacy struct {
	Text string `json:"text"`
}
