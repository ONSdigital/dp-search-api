package elasticsearch

import _ "embed"

//go:embed mappings.json
var mappingsJson []byte

func GetDefaultMappings() []byte {
	return mappingsJson
}
