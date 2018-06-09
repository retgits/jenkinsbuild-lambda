#!/bin/bash

# Get the dependencies
deps() {
   go get -u ./... 
}

# Remove the bin folder
clean() {
   rm -rf ./bin 
}

# Create the executable
build() {
    GOOS=linux GOARCH=amd64 go build -o ./bin/main main.go
}

case "$1" in
    "deps")
        deps
        ;;
    "clean")
        clean
        ;;
    "build")
        build
        ;;
    *)
        echo "The target {$1} want to execute doesn't exist"
        echo 
        echo "Usage"
        echo "./build deps      : go get and update all the dependencies"
        echo "./build clean     : removes the ./bin folder"
        echo "./build build     : creates the executable"
        exit 2
        ;; 
esac