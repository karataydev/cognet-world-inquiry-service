PROJECT_NAME=$(notdir $(patsubst %/,%,$(CURDIR)))

build:
	@go build -o bin/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)

run: build
	@./bin/$(PROJECT_NAME) $(ARGS)

test:
	@go test -v ./..

compile:
	GOOS=freebsd GOARCH=386 go build -o bin/$(PROJECT_NAME)-freebsd-386 main.go
	GOOS=linux GOARCH=386 go build -o bin/$(PROJECT_NAME)-linux-386 main.go
	GOOS=windows GOARCH=386 go build -o bin/$(PROJECT_NAME)-windows-386 main.go

build-all: windows linux darwin
	@echo version: $(VERSION)

EXECUTABLE=$(PROJECT_NAME)
VERSION=$(shell git describe --tags)
WINDOWS=$(EXECUTABLE)_windows_amd64_$(VERSION).exe
LINUX=$(EXECUTABLE)_linux_amd64_$(VERSION)
DARWIN=$(EXECUTABLE)_darwin_amd64_$(VERSION)

windows: $(WINDOWS)

linux: $(LINUX)

darwin: $(DARWIN)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -v -o bin/$(WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/$(PROJECT_NAME)/main.go

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -v -o bin/$(LINUX) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/$(PROJECT_NAME)/main.go

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -v -o bin/$(DARWIN) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/$(PROJECT_NAME)/main.go

clean:
	rm -f $(WINDOWS) $(LINUX) $(DARWIN)
