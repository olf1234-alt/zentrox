.PHONY: fmt lint test build examples

fmt:
	gofmt -s -w .
	go mod tidy

lint:
	golangci-lint run	

test:
	go test ./... -race -count=1

build:
	go build ./...

examples:
	for d in ./examples/*; do if [ -d "$$d" ]; then (cd "$$d" && go run . || true); fi; done
	
