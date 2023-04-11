package models

type Hub struct {
	Scrubber Scrubber
	Category Category
	Berlin   Berlin
}

type Scrubber struct {
	Query   string  `json:"query"`
	Results Results `json:"results,omitempty"`
	Time    string  `json:"time,omitempty"`
}

type Results struct {
	Areas      []AreaResp     `json:"areas,omitempty"`
	Industries []IndustryResp `json:"industries,omitempty"`
}

type AreaResp struct {
	Codes      map[string]string `json:"codes,omitempty"`
	Name       string            `json:"name,omitempty"`
	Region     string            `json:"region,omitempty"`
	RegionCode string            `json:"region_code,omitempty"`
}

type IndustryResp struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type Category []struct {
	Code  []string `json:"c,omitempty"`
	Score float32  `json:"s,omitempty"`
}

type Berlin struct {
	Query   SearchTermJson `json:"query,omitempty"`
	Results []SearchResult `json:"results,omitempty"`
	Time    string         `json:"time,omitempty"`
}

type SearchTermJson struct {
	Codes           []string    `json:"codes,omitempty"`
	ExactMatches    []string    `json:"exact_matches,omitempty"`
	Normalized      string      `json:"normalized"`
	NotExactMatches []string    `json:"not_exact_matches,omitempty"`
	Raw             string      `json:"raw,omitempty"`
	StateFilter     interface{} `json:"state_filter,omitempty"`
	StopWords       []string    `json:"stop_words,omitempty"`
}

type SearchResult struct {
	Loc   LocJson `json:"loc,omitempty"`
	Score int     `json:"score,omitempty"`
}

type LocJson struct {
	Codes    []string    `json:"codes,omitempty"`
	Encoding string      `json:"encoding,omitempty"`
	Id       string      `json:"id,omitempty"`
	Key      string      `json:"key,omitempty"`
	Names    []string    `json:"names,omitempty"`
	State    []string    `json:"state,omitempty"`
	Subdiv   interface{} `json:"subdiv,omitempty"`
}
