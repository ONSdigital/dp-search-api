#!/bin/bash
COUNTER=0
if [ -z $ELASTIC_URL ]; then
    ELASTIC_URL=localhost:9200
fi
printf "waiting for $ELASTIC_URL   "
until $(curl --output /dev/null --silent --head --fail $ELASTIC_URL); do
    printf '.'
    sleep 5
    if [ $COUNTER -gt 10 ]; then
        printf "\n waited too long"
        exit
    fi
    let COUNTER=COUNTER+1
done