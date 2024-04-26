module github.com/ONSdigital/dp-search-api

go 1.22

// to fix: [CVE-2023-32731] CWE-Other
replace google.golang.org/grpc => google.golang.org/grpc v1.55.0

// to avoid the following vulnerabilities:
//     - CVE-2022-29153 # pkg:golang/github.com/hashicorp/consul/api@v1.1.0 and pkg:golang/github.com/hashicorp/consul/sdk@v0.1.1
//     - sonatype-2021-1401 # pkg:golang/github.com/miekg/dns@v1.0.14
//     - sonatype-2019-0890 # pkg:golang/github.com/pkg/sftp@v1.10.1
replace github.com/spf13/cobra => github.com/spf13/cobra v1.7.0

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.260.0
	github.com/ONSdigital/dp-authorisation v0.2.1
	github.com/ONSdigital/dp-component-test v0.11.0
	github.com/ONSdigital/dp-elasticsearch/v3 v3.0.1-alpha.4.0.20230308115225-bb7559a89d0c
	github.com/ONSdigital/dp-healthcheck v1.6.3
	github.com/ONSdigital/dp-net/v2 v2.11.2
	github.com/ONSdigital/dp-otel-go v0.0.7
	github.com/ONSdigital/dp-search-scrubber-api v0.4.1
	github.com/ONSdigital/log.go/v2 v2.4.3
	github.com/cucumber/godog v0.14.0
	github.com/elastic/go-elasticsearch/v7 v7.17.10
	github.com/google/go-cmp v0.6.0
	github.com/gorilla/mux v1.8.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/maxcnunes/httpfake v1.2.4
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.8.1
	github.com/tdewolff/minify v2.3.6+incompatible
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.51.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.51.0
)

require (
	github.com/ONSdigital/dp-mongodb-in-memory v1.7.0 // indirect
	github.com/ONSdigital/dp-rchttp v1.0.0 // indirect
	github.com/ONSdigital/go-ns v0.0.0-20210916104633-ac1c1c52327e // indirect
	github.com/aws/aws-sdk-go v1.51.29 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/chromedp/cdproto v0.0.0-20240421230201-ab917191657d // indirect
	github.com/chromedp/chromedp v0.9.5 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/cucumber/gherkin/go/v26 v26.2.0 // indirect
	github.com/cucumber/messages/go/v21 v21.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.3.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smarty/assertions v1.15.1 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/tdewolff/test v1.0.6 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240424034433-3c2c7870ae76 // indirect
	go.mongodb.org/mongo-driver v1.15.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.51.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.26.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.26.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.26.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.26.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.26.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/otel/sdk v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
