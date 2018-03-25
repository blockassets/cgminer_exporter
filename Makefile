DATE=$(shell date -u '+%Y-%m-%d %H:%M:%S')
COMMIT=$(shell git log --format=%h -1)
VERSION=main.version=${TRAVIS_BUILD_NUMBER} ${COMMIT} ${DATE}
COMPILE_FLAGS=-ldflags="-X '${VERSION}'"

build:
	@go build ${COMPILE_FLAGS}

arm:
	@GOOS=linux GOARCH=arm GOARM=7 go build ${COMPILE_FLAGS}

dep:
	@dep ensure

test:
	@go test ./...

clean:
	@rm -f cgminer_exporter

all: clean test build
