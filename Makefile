.PHONY: build test format lint lint-fix build-go build-artifacts

LAMBDA_SERVICES := create delete retrieve

build-go:
	@mkdir -p bin
	@for service in ${LAMBDA_SERVICES}; do \
		echo "Building Service: $$service"; \
		GOFLAGS=-buildvcs=false GOOS=linux go build -ldflags="-s -w" -o "bin/$$service/bootstrap" "./cmd/$$service"; \
		done


build-artifacts: 
	$(MAKE) build-go
	@for service in ${LAMBDA_SERVICES}; do \	
		echo "Packing Artifacts: $$service"; \
			cd bin && zip -j $$service.zip $$service/bootstrap && rm -r $$service; \	
		done

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

