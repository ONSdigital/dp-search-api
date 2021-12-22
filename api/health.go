package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
)

type healthMessage struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func HealthCheckHandlerCreator(elasticSearchClient ElasticSearcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		var (
			healthIssue string
			err         error
		)

		// assume all well
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		body := []byte("{\"status\":\"OK\"}") // quicker than json.Marshal(healthMessage{...})

		// test elastic access
		res, err := elasticSearchClient.GetStatus(ctx)
		if err != nil {
			healthIssue = err.Error()
		} else if !isElasticSearchHealthy(string(res)) {
			healthIssue = string(res)
		}

		// when there's a healthIssue, change headers and content
		if healthIssue != "" {
			w.WriteHeader(http.StatusInternalServerError)
			if body, err = json.Marshal(healthMessage{
				Status: "error",
				Error:  healthIssue,
			}); err != nil {
				log.Error(ctx, "elasticsearch healthcheck status json failed to parse", err)
				panic(err)
			}
		}

		// return json
		fmt.Fprint(w, string(body))
	}
}

func isElasticSearchHealthy(res string) bool {
	if strings.Contains(res, " green ") {
		return true
	}

	if strings.Contains(res, " yellow ") {
		return true
	}

	return false
}
