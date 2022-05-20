package models

// *************************************************************
// structs representing the raw elastic search response
// *************************************************************

// EsResponse holds a response slice from ES
type EsResponses struct {
	Responses []EsResponse `json:"responses"`
}

type EsResponse struct {
	Took         int                    `json:"took"`
	Hits         ESResponseHits         `json:"hits"`
	Aggregations ESResponseAggregations `json:"aggregations"`
	Suggest      Suggest                `json:"suggest"`
}

type Suggest struct {
	SearchSuggest []SearchSuggest `json:"search_suggest"`
}

type SearchSuggest struct {
	Text    string   `json:"text"`
	Offset  int      `json:"offset"`
	Length  int      `json:"length"`
	Options []Option `json:"options"`
}

type Option struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}

type ESResponseHits struct {
	Hits []ESResponseHit `json:"hits"`
}

type ESResponseHit struct {
	Source    ESSourceDocument `json:"_source"`
	Highlight ESHighlight      `json:"highlight"`
}

type ESResponseAggregations struct {
	DocCounts ESDocCounts `json:"docCounts"`
}

type ESDocCounts struct {
	Buckets []ESBucket `json:"buckets"`
}

type ESBucket struct {
	Key   string `json:"key"`
	Count int    `json:"doc_count"`
}

type ESSourceDocument struct {
	DataType        string   `json:"type"`
	CDID            string   `json:"cdid"`
	DatasetID       string   `json:"dataset_id"`
	Keywords        []string `json:"keywords"`
	MetaDescription string   `json:"meta_description"`
	ReleaseDate     string   `json:"release_date,omitempty"`
	Summary         string   `json:"summary"`
	Title           string   `json:"title"`
	Topics          []string `json:"topics"`
	URI             string   `json:"uri"`
}

type ESHighlight struct {
	DescriptionTitle     []*string `json:"description.title"`
	DescriptionEdition   []*string `json:"description.edition"`
	DescriptionSummary   []*string `json:"description.summary"`
	DescriptionMeta      []*string `json:"description.metaDescription"`
	DescriptionKeywords  []*string `json:"description.keywords"`
	DescriptionDatasetID []*string `json:"description.datasetId"`
}

// ********************************************************
// Structs representing the transformed response
// ********************************************************

type SearchResponse struct {
	Took                int                `json:"took"`
	ContentTypes        []ContentType      `json:"content_types"`
	Items               []ESSourceDocument `json:"items"`
	Suggestions         []string           `json:"suggestions,omitempty"`
	AdditionSuggestions []string           `json:"additional_suggestions,omitempty"`
}
