.PHONY: test
test:
	@go test -v -race -failfast -count 1 -coverprofile cover.out ./...

.PHONY: bench
bench:
	@go test -v -count 1 -run x -bench . ./...

.PHONY: lint
lint: go.sum
	@golangci-lint run --new=false ./...
	
go.sum: go.mod
	@go mod tidy
