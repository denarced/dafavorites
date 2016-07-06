#!/bin/bash

while [ 1 ]
do
    clear
    date
    goFiles="dafavorites.go deviantart/fetch.go"
    echo "--- go test"
    go test ./...
    echo "--- go lint"
    golint ./... | egrep -v -f goimports_patterns.txt
    echo "--- go vet"
    go vet .
    echo "--- gotags"
    gotags -R -f tags .
    echo "--- wait for file changes"
    inotifywait -e close_write $goFiles
done
