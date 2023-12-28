tidy:
	@go mod tidy

build:
	@tidy
	@go build -ldflags "-s -w"

checksum:
	@sha256sum -b plugin-template-go

update-all:
	@go get -u ./...

test:
	@go test -v ./...

build-dev: tidy
	@CGO_ENABLED=0 go build
