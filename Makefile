all: slow-lint lint

slow-lint:
	@golint .

lint:
	@golangci-lint run --enable-all

deps:
	@go get -u golang.org/x/lint/golint
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.21.0
