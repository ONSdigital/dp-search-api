package models

// *************************************************************
// structs representing the raw elastic search response
// *************************************************************

// EsResponse holds a response slice from ES
type EsResponses struct {
	Responses []*EsResponse `json:"responses"`
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
	Total int
	Hits  []ESResponseHit `json:"hits"`
}

type ESResponseHit struct {
	Source    ESSourceDocument `json:"_source"`
	Highlight *ESHighlight     `json:"highlight"`
}

type ESResponseAggregations struct {
	ContentTypes       ESDocCounts `json:"content_types"`
	Topic              ESDocCounts `json:"topic"`
	PopulationType     ESDocCounts `json:"population_type"`
	Dimensions         ESDocCounts `json:"dimensions"`
	DistinctTopicCount CountValue  `json:"distinct_topics_count"`
}

type ESDocCounts struct {
	Buckets []ESBucket `json:"buckets"`
}

type CountValue struct {
	Value int `json:"value"`
}

type ESBucket struct {
	Key   string `json:"key"`
	Count int    `json:"doc_count"`
}

type ESSourceDocument struct {
	DataType        string              `json:"type"`
	CDID            string              `json:"cdid"`
	DatasetID       string              `json:"dataset_id"`
	Edition         string              `json:"edition"`
	Keywords        []string            `json:"keywords"`
	MetaDescription string              `json:"meta_description"`
	ReleaseDate     string              `json:"release_date,omitempty"`
	Summary         string              `json:"summary"`
	Title           string              `json:"title"`
	Topics          []string            `json:"topics"`
	URI             string              `json:"uri"`
	Highlight       *HighlightObj       `json:"highlight,omitempty"`
	DateChanges     []ReleaseDateChange `json:"date_changes,omitempty"`
	Cancelled       bool                `json:"cancelled,omitempty"`
	CanonicalTopic  string              `json:"canonical_topic"`
	Finalised       bool                `json:"finalised,omitempty"`
	ProvisionalDate string              `json:"provisional_date,omitempty"`
	Published       bool                `json:"published,omitempty"`
	Language        string              `json:"language,omitempty"`
	Survey          string              `json:"survey,omitempty"`
	PopulationType  ESPopulationType    `json:"population_type,omitempty"`
	Dimensions      []ESDimensions      `json:"dimensions,omitempty"`
}

type HighlightObj struct {
	DatasetID       string    `json:"dataset_id,omitempty"`
	Keywords        []*string `json:"keywords,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	Summary         string    `json:"summary,omitempty"`
	Title           string    `json:"title,omitempty"`
}

type ESHighlight struct {
	Title     []*string `json:"title"`
	Edition   []*string `json:"edition"`
	Summary   []*string `json:"summary"`
	MetaDesc  []*string `json:"meta_description"`
	Keywords  []*string `json:"keywords"`
	DatasetID []*string `json:"dataset_id"`
}

// TopicCount represents the API response for an aggregation of topics
type TopicCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// ContentTypeCount represents the API response for an aggregation of content types
type ContentTypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// DimensionCount represents the API response for an aggregation of dimensions
type DimensionCount struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// PopulationTypeCount represents the API response for an aggregation of population types
type PopulationTypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type ESPopulationType struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type ESDimensions struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	RawLabel string `json:"raw_label"`
}

// ********************************************************
// Structs representing the transformed response
// ********************************************************

type Item struct {
	DataType        string              `json:"type"`
	CDID            string              `json:"cdid"`
	DatasetID       string              `json:"dataset_id"`
	Edition         string              `json:"edition"`
	Keywords        []string            `json:"keywords"`
	MetaDescription string              `json:"meta_description"`
	ReleaseDate     string              `json:"release_date,omitempty"`
	Summary         string              `json:"summary"`
	Title           string              `json:"title"`
	Topics          []string            `json:"topics"`
	URI             string              `json:"uri"`
	Highlight       *HighlightObj       `json:"highlight,omitempty"`
	DateChanges     []ReleaseDateChange `json:"date_changes,omitempty"`
	Cancelled       bool                `json:"cancelled,omitempty"`
	CanonicalTopic  string              `json:"canonical_topic"`
	Finalised       bool                `json:"finalised,omitempty"`
	ProvisionalDate string              `json:"provisional_date,omitempty"`
	Published       bool                `json:"published,omitempty"`
	Language        string              `json:"language,omitempty"`
	Survey          string              `json:"survey,omitempty"`
	PopulationType  string              `json:"population_type,omitempty"`
	Dimensions      []ESDimensions      `json:"dimensions,omitempty"`
}

type SearchResponse struct {
	Count               int                   `json:"count"`
	Took                int                   `json:"took"`
	DistinctItemsCount  int                   `json:"distinct_items_count"`
	Topics              []TopicCount          `json:"topics"`
	ContentTypes        []ContentTypeCount    `json:"content_types"`
	Items               []Item                `json:"items"`
	Suggestions         []string              `json:"suggestions,omitempty"`
	AdditionSuggestions []string              `json:"additional_suggestions,omitempty"`
	Dimensions          []DimensionCount      `json:"dimensions,omitempty"`
	PopulationType      []PopulationTypeCount `json:"population_type,omitempty"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateChange struct {
	ChangeNotice string `json:"change_notice"`
	Date         string `json:"previous_date"`
}
