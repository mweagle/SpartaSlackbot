.DEFAULT_GOAL=build
.PHONY: build test get tags clean

ensure_vendor:
	mkdir -pv vendor

clean:
	rm -rf ./vendor
	go clean .

get:
	go get -u github.com/golang/dep/...
	dep ensure

generate:
	go generate -x

format:
	go fmt .

vet: generate
	go vet .

build: format generate vet
	go build .
	@echo "Build complete"

test:
	go test -v .

tags:
	gotags -tag-relative=true -R=true -sort=true -f="tags" -fields=+l .

provision: build
	go run slack.go --level info provision --s3Bucket $(S3_BUCKET)

delete:
	go run slack.go delete

describe: build
	rm -rf ./graph.html
	go run slack.go describe
