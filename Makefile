export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

BUILD_ARCH=build/$(GOOS)-$(GOARCH)
HASH?=$(shell make hash)
DATE:=$(shell date '+%Y%m%d-%H%M%S')
TGZ_FILE=dp-search-query-$(GOOS)-$(GOARCH)-$(DATE)-$(HASH).tar.gz

build:
	@mkdir -p $(BUILD_ARCH)/bin
	go build -o $(BUILD_ARCH)/bin/dp-search-query cmd/dp-search-query/main.go
	cp -r templates $(BUILD_ARCH)/templates

package: build
	tar -zcf $(TGZ_FILE) -C $(BUILD_ARCH) .

nomad:
	@healthcheck_port=$${HEALTHCHECK_ADDR#*:};  \
	sed	-e 's,DATA_CENTER,$(DATA_CENTER),g' \
		-e 's,S3_TAR_FILE,$(S3_TAR_FILE),g' \
		-e 's,ELASTIC_SEARCH_URL,$(ELASTIC_URL),g' \
		-e 's,HEALTHCHECK_PORT,'$$healthcheck_port',g' \
		-e 's,HEALTHCHECK_ENDPOINT,$(HEALTHCHECK_ENDPOINT),g' \
		< dp-search-query-template.nomad > dp-search-query.nomad
hash:
	@git rev-parse --short HEAD

debug: build
	HUMAN_LOG=1 ./build/dp-search-query

waitOnElastic:   export ELASTIC_URL = http://localhost:9999/
waitOnElastic:
	pause.sh


test: 	export BIND_ADDR = :10002
test:   export ELASTIC_URL = http://localhost:9999/
test:
	./waitForElastic.sh
	go test

bdd: startElastic test stopElastic


startElastic:
	docker run --name es-bdd  -d   -p 9999:9200 guidof/onswebsite-search:5.0.0

stopElastic:
	docker rm -f es-bdd

.PHONY: build debug test startElastic stopElastic package
