package models

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
	Description Description `json:"description"`
	Type        string      `json:"type"`
	URI         string      `json:"uri"`
}

type Description struct {
	Contact           *contact      `json:"contact,omitempty"`
	DatasetID         string        `json:"dataset_id,omitempty"`
	Edition           string        `json:"edition,omitempty"`
	Headline1         string        `json:"headline1,omitempty"`
	Headline2         string        `json:"headline2,omitempty"`
	Headline3         string        `json:"headline3,omitempty"`
	Highlight         *HighlightObj `json:"highlight,omitempty"`
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

type HighlightObj struct {
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
