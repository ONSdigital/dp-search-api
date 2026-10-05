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
	PreviousURIs    []string            `json:"previous_uris,omitempty"`
	Published       bool                `json:"published,omitempty"`
	Language        string              `json:"language,omitempty"`
	Survey          string              `json:"survey,omitempty"`
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

// FilterCount represents the API response for an aggregation
type FilterCount struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Count int    `json:"count"`
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
	PreviousURIs    []string            `json:"previous_uris,omitempty"`
	Published       bool                `json:"published,omitempty"`
	Language        string              `json:"language,omitempty"`
	Survey          string              `json:"survey,omitempty"`
	Dimensions      []ESDimensions      `json:"dimensions,omitempty"`
}

type SearchResponse struct {
	Count               int           `json:"count"`
	Took                int           `json:"took"`
	DistinctItemsCount  int           `json:"distinct_items_count"`
	Topics              []FilterCount `json:"topics"`
	ContentTypes        []FilterCount `json:"content_types"`
	Items               []Item        `json:"items"`
	Suggestions         []string      `json:"suggestions,omitempty"`
	AdditionSuggestions []string      `json:"additional_suggestions,omitempty"`
	Dimensions          []FilterCount `json:"dimensions,omitempty"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateChange struct {
	ChangeNotice string `json:"change_notice"`
	Date         string `json:"previous_date"`
}
