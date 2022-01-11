package elasticsearch

import _ "embed"

//go:embed search-index-settings.json
var searchIndexSettingsJSON []byte

func GetSearchIndexSettings() []byte {
	return searchIndexSettingsJSON
}
