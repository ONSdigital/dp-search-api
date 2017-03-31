.PHONY: build debug test startElastic stopElastic
build:
	go build -o build/dp-search-query

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