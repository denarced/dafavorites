#!/bin/bash

waitForChange() {
    inotifywait -q -e close_write `find cmd/ shared/ -type f -name \*.go`
}

green() {
    echo -e "\e[32m-- $1\e[0m"
}

if [ "$1" == "-h" ] || [ "$1" == "--help" ]
then
    echo "Usage: $0 [-h|--help] [-c|--html-coverage]"
    exit 1
fi

coverOption="-cover"
coverageFile=""
if [ "$1" == "-c" ] || [ "$1" == "--html-coverage" ]
then
    # value doesn't matter
    coverOption="-coverprofile"
    coverageFile="cover.out"
fi

while :
do
    clear
    green "BUILD"
    go install ./...
    if [ $? -ne 0 ]
    then
        waitForChange
        continue
    fi

    green "TEST"
    go test $coverOption $coverageFile -covermode count ./...
    if [ $? -ne 0 ]
    then
        waitForChange
        continue
    fi

    if [ -n "$coverageFile" ]
    then
        go tool cover -html=$coverageFile
        if [ $? -ne 0 ]
        then
            waitForChange
            continue
        fi
    fi

    green "WAIT"
    waitForChange
done
