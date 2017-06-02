package utils

import (
	"net/http"
	"os"
)

func GetEnvironmentVariable(name string, defaultValue string) string {
	environmentValue := os.Getenv(name)
	if environmentValue != "" {
		return environmentValue
	}
	return defaultValue
}

func SetCSVContentHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Disposition", "attachment; filename=data.csv")
	w.Header().Set("Content-Type", "text/csv")
}

func SetXLSContentHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Disposition", "attachment; filename=data.xls")
	w.Header().Set("Content-Type", "application/vnd.ms-excel")
}
