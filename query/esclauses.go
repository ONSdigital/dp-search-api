package query

import (
	"fmt"
	"time"
)

const (
	Separator = ","

	NoProvisionalNoConfirmedPostponed = `
{"term":{"description.finalised":true}}, {"exists":{"field":"dateChanges"}}
`

	NoProvisionalConfirmedNoPostponed = `
{"term":{"description.finalised":true}}, {"bool":{"must_not":{"exists":{"field":"dateChanges"}}}}
`

	NoProvisionalConfirmedPostponed = `
{"term":{"description.finalised":true}}
`

	ProvisionalNoConfirmedNoPostponed = `
{"term":{"description.finalised":false}}
`

	ProvisionalNoConfirmedPostponed = `
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

	ProvisionalConfirmedNoPostponed = `
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
	return fmt.Sprintf("%s, %s, %s", `{"term": {"description.published": false}}`, `{"term": {"description.cancelled": false}}`,
		fmt.Sprintf(`{"range": {"description.releaseDate": {"gte": %q}}}`, now.Format(dateFormat)))
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
