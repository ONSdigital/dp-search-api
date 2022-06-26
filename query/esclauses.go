package query

import (
	"fmt"
	"time"
)

const (
	Separator   = ","
	EmptyClause = `{}`

	NoProvisionalNoConfirmedPostponed = `
{"term":{"finalised":true}}, {"exists":{"field":"date_changes"}}
`

	NoProvisionalConfirmedNoPostponed = `
{"term":{"finalised":true}}, {"bool":{"must_not":{"exists":{"field":"date_changes"}}}}
`

	NoProvisionalConfirmedPostponed = `
{"term":{"finalised":true}}
`

	ProvisionalNoConfirmedNoPostponed = `
{"term":{"finalised":false}}
`

	ProvisionalNoConfirmedPostponed = `
{"bool":{
    "should":[
      {"term":{"finalised":false}},
      {"bool":{
          "must":[
            {"term":{"finalised":true}},
            {"exists":{"field":"date_changes"}}
          ]}
      }
    ]}
}`

	ProvisionalConfirmedNoPostponed = `
{"bool":{
    "should":[
      {"term":{"finalised":false}},
      {"bool":{
          "must":[
            {"term":{"finalised":true}},
            {"bool":{"must_not":{"exists":{"field":"date_changes"}}}}
          ]}
      }
    ]}
}`

	LegacyNoProvisionalNoConfirmedPostponed = `
{"term":{"description.finalised":true}}, {"exists":{"field":"dateChanges"}}
`

	LegacyNoProvisionalConfirmedNoPostponed = `
{"term":{"description.finalised":true}}, {"bool":{"must_not":{"exists":{"field":"dateChanges"}}}}
`

	LegacyNoProvisionalConfirmedPostponed = `
{"term":{"description.finalised":true}}
`

	LegacyProvisionalNoConfirmedNoPostponed = `
{"term":{"description.finalised":false}}
`

	LegacyProvisionalNoConfirmedPostponed = `
{"bool":{
    "should":[
      {"term":{"description.finalised":false}},
      {"bool":{
          "must":[
            {"term":{"description.finalised":true}},
            {"exists":{"field":"dateChanges"}}
          ]}
      }
    ]}
}`

	LegacyProvisionalConfirmedNoPostponed = `
{"bool":{
    "should":[
      {"term":{"description.finalised":false}},
      {"bool":{
          "must":[
            {"term":{"description.finalised":true}},
            {"bool":{"must_not":{"exists":{"field":"dateChanges"}}}}
          ]}
      }
    ]}
}`
)

func mainUpcomingClause(now time.Time) string {
	return fmt.Sprintf("%s, %s, %s", `{"term": {"published": false}}`, `{"term": {"cancelled": false}}`,
		fmt.Sprintf(`{"range": {"release_date": {"gte": %q}}}`, now.Format(dateFormat)))
}

func supplementaryUpcomingClause(sr ReleaseSearchRequest) string {
	switch {
	case !sr.Provisional && !sr.Confirmed && sr.Postponed:
		return NoProvisionalNoConfirmedPostponed
	case !sr.Provisional && sr.Confirmed && !sr.Postponed:
		return NoProvisionalConfirmedNoPostponed
	case !sr.Provisional && sr.Confirmed && sr.Postponed:
		return NoProvisionalConfirmedPostponed
	case sr.Provisional && !sr.Confirmed && !sr.Postponed:
		return ProvisionalNoConfirmedNoPostponed
	case sr.Provisional && !sr.Confirmed && sr.Postponed:
		return ProvisionalNoConfirmedPostponed
	case sr.Provisional && sr.Confirmed && !sr.Postponed:
		return ProvisionalConfirmedNoPostponed
	}

	return ""
}

func legacyMainUpcomingClause(now time.Time) string {
	return fmt.Sprintf("%s, %s, %s", `{"term": {"description.published": false}}`, `{"term": {"description.cancelled": false}}`,
		fmt.Sprintf(`{"range": {"description.releaseDate": {"gte": %q}}}`, now.Format(dateFormat)))
}

func legacySupplementaryUpcomingClause(sr LegacyReleaseSearchRequest) string {
	switch {
	case !sr.Provisional && !sr.Confirmed && sr.Postponed:
		return LegacyNoProvisionalNoConfirmedPostponed
	case !sr.Provisional && sr.Confirmed && !sr.Postponed:
		return LegacyNoProvisionalConfirmedNoPostponed
	case !sr.Provisional && sr.Confirmed && sr.Postponed:
		return LegacyNoProvisionalConfirmedPostponed
	case sr.Provisional && !sr.Confirmed && !sr.Postponed:
		return LegacyProvisionalNoConfirmedNoPostponed
	case sr.Provisional && !sr.Confirmed && sr.Postponed:
		return LegacyProvisionalNoConfirmedPostponed
	case sr.Provisional && sr.Confirmed && !sr.Postponed:
		return LegacyProvisionalConfirmedNoPostponed
	}

	return ""
}
