.PHONY: build test format lint lint-fix build-go build-artifacts deploy update

build-go:
	@echo "Building service"
	mkdir -p bin/cmd-cv
	GOFLAGS=-buildvcs=false GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/cmd-cv/bootstrap ./cmd/

build-artifacts: build-go
	@echo "Packing artifacts"
	cd bin/cmd-cv && zip -j -FS artifact.zip bootstrap

test:
	@echo "Testing"
	go test -timeout 30s ./...

format:
	@echo "Tidying"
	go mod tidy
	@echo "Formatting"
	go fmt ./...

lint:
	GOFLAGS=-buildvcs=false go tool golangci-lint run

lint-fix:
	GOFLAGS=-buildvcs=false go tool golangci-lint run --fix

deploy:
	sam deploy \
		--stack-name cmd-cv \
		--region us-west-2 \
		--capabilities CAPABILITY_IAM \
		--no-confirm-changeset

update:
	@echo "Updating"
	go get -u ./...
	$(MAKE) format
