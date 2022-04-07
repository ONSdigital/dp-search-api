package models

// *************************************************************
// structs representing the raw elastic search response
// *************************************************************

// Es7xResponse holds a response slice from ES
type Es7xResponse struct {
	Responses []EsResponse7x `json:"responses"`
}

type EsResponse7x struct {
	Took         int                      `json:"took"`
	Hits         ES7xResponseHits         `json:"hits"`
	Aggregations ES7xResponseAggregations `json:"aggregations"`
	Suggest      []string                 `json:"suggest"`
}

type ES7xResponseHits struct {
	Total int
	Hits  []ES7xResponseHit `json:"hits"`
}

type ES7xResponseHit struct {
	Source    []ES7xSourceDocument `json:"_source"`
	Highlight ES7xHighlight        `json:"highlight"`
}

type ES7xResponseAggregations struct {
	Doccounts ES7xDocCounts `json:"docCounts"`
}

type ES7xDocCounts struct {
	Buckets []ES7xBucket `json:"buckets"`
}

type ES7xBucket struct {
	Key   string `json:"key"`
	Count int    `json:"doc_count"`
}

type ES7xSourceDocument struct {
	DataType        string   `json:"type"`
	JobID           string   `json:"job_id"`
	SearchIndex     string   `json:"search_index"`
	CDID            string   `json:"cdid"`
	DatasetID       string   `json:"dataset_id"`
	Keywords        []string `json:"keywords"`
	MetaDescription string   `json:"meta_description"`
	ReleaseDate     string   `json:"release_date,omitempty"`
	Summary         string   `json:"summary"`
	Title           string   `json:"title"`
	Topics          []string `json:"topics"`
}

type ES7xHighlight struct {
	DescriptionTitle     *[]string `json:"description.title"`
	DescriptionEdition   *[]string `json:"description.edition"`
	DescriptionSummary   *[]string `json:"description.summary"`
	DescriptionMeta      *[]string `json:"description.metaDescription"`
	DescriptionKeywords  *[]string `json:"description.keywords"`
	DescriptionDatasetID *[]string `json:"description.datasetId"`
}

// ********************************************************
// Structs representing the transformed response
// ********************************************************

type Search7xResponse struct {
	Count               int                  `json:"count"`
	Took                int                  `json:"took"`
	Items               []ES7xSourceDocument `json:"items"`
	Suggestions         []string             `json:"suggestions,omitempty"`
	AdditionSuggestions []string             `json:"additional_suggestions,omitempty"`
}
