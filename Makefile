.PHONY: build test format lint lint-fix build-go build-artifacts

build-go:
	echo "Building Service\n";
	mkdir -p bin/cmd-cv
	GOFLAGS=-buildvcs=false GOOS=linux go build -ldflags="-s -w" -o "bin/cmd-cv/bootstrap" "./cmd/"


build-artifacts: 
	echo "Packing Artifacts";
	$(MAKE) build-go
	cd bin && zip -j artifact.zip cmd-cv/bootstrap && rm -r cmd-cv/bootstrap

test:
	@echo "\t Testing"
	@go test -timeout 30s ./...


format:
	@echo "\t Tidying"
	@go mod tidy
	
	@echo "\t Formatting"
	@go fmt ./...


lint:
	GOFLAGS=-buildvcs=false go tool golangci-lint run


lint-fix:
	GOFLAGS=-buildvcs=false go tool golangci-lint run --fix


update:
	@echo "\t Updating"
	@go get -u ./... && $(MAKE) format
