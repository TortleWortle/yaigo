.PHONY: test lint

test: lint
	go test ./...

lint:
	golangci-lint run
