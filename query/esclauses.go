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
