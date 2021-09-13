package elasticsearch

import _ "embed"

//go:embed search-index-settings.json
var searchIndexSettingsJson []byte

func GetSearchIndexSettings() []byte {
	return searchIndexSettingsJson
}
