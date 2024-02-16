.PHONY: build format fmt run tidy deps

build: build/stream-inator
build/stream-inator:
	@go build -o "$@" .

build-mac:
build/stream-inator-mac:
	@GOOS=darwin GOARCH=arm64 go build -o "$@" .

build-windows:
build/stream-inator-windows.exe:
	@GOOS=windows GOARCH=x86_64 go build -o "$@" -ldflags "-H=windowsgui" .

run:
	@go run .

tidy: deps
deps:
	@go mod tidy
	@go mod download

format: fmt
fmt:
	@gofmt -w -s *.go
