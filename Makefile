export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

export S3_TAR_FILE?=$(shell go env S3_TAR_FILE)
export ELASTIC_URL?=$(shell go env ELASTIC_URL)
export DATA_CENTER?=$(shell go env DATA_CENTER)

BUILD_ARCH=build/$(GOOS)-$(GOARCH)
HASH?=$(shell make hash)
DATE:=$(shell date '+%Y%m%d-%H%M%S')
TGZ_FILE=dp-search-query-$(GOOS)-$(GOARCH)-$(DATE)-$(HASH).tar.gz

build:
	@mkdir -p $(BUILD_ARCH)
	go build -o $(BUILD_ARCH)/bin/dp-search-query cmd/dp-search-query/main.go
	cp -r templates $(BUILD_ARCH)/templates
	
package: build
	tar -zcf $(TGZ_FILE) -C $(BUILD_ARCH) .
	
nomad:
	@cp dp-search-query-template.nomad dp-search-query.nomad
	@sed -i.bak s,DATA_CENTER,$(DATA_CENTER),g dp-search-query.nomad
	@sed -i.bak s,S3_TAR_FILE,$(S3_TAR_FILE),g dp-search-query.nomad
	@sed -i.bak s,ELASTIC_SEARCH_URL,$(ELASTIC_URL),g dp-search-query.nomad	
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