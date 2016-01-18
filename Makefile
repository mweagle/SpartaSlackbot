.DEFAULT_GOAL=build
.PHONY: build test get tags clean

ensure_vendor:
	mkdir -pv vendor

clean:
	rm -rf ./vendor
	go clean .

get: clean ensure_vendor
	git clone --depth=1 https://github.com/aws/aws-sdk-go ./vendor/github.com/aws/aws-sdk-go
	rm -rf ./src/main/vendor/github.com/aws/aws-sdk-go/.git
	git clone --depth=1 https://github.com/vaughan0/go-ini ./vendor/github.com/vaughan0/go-ini
	rm -rf ./src/main/vendor/github.com/vaughan0/go-ini/.git
	git clone --depth=1 https://github.com/Sirupsen/logrus ./vendor/github.com/Sirupsen/logrus
	rm -rf ./src/main/vendor/github.com/Sirupsen/logrus/.git
	git clone --depth=1 https://github.com/voxelbrain/goptions ./vendor/github.com/voxelbrain/goptions
	rm -rf ./src/main/vendor/github.com/voxelbrain/goptions/.git
	git clone --depth=1 https://github.com/mjibson/esc ./vendor/github.com/mjibson/esc
	rm -rf ./src/main/vendor/github.com/mjibson/esc/.git
	git clone --depth=1 https://github.com/mweagle/Sparta ./vendor/github.com/mweagle/Sparta
	rm -rf ./src/main/vendor/github.com/mweagle/Sparta/.git
	git clone --depth=1 https://github.com/mweagle/go-cloudformation ./vendor/github.com/mweagle/go-cloudformation
	rm -rf ./src/main/vendor/github.com/mweagle/go-cloudformation/.git

generate:
	go generate -x

format:
	go fmt .

vet: generate
	go vet .

build: format generate vet
	GO15VENDOREXPERIMENT=1 go build .
	@echo "Build complete"

test:
	GO15VENDOREXPERIMENT=1 go test -v .

tags:
	gotags -tag-relative=true -R=true -sort=true -f="tags" -fields=+l .

provision: build
	go run slack.go --level info provision --s3Bucket $(S3_BUCKET)

delete:
	GO15VENDOREXPERIMENT=1 go run slack.go delete

describe: build
	rm -rf ./graph.html
	go run slack.go describe
