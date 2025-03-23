.PHONY: goimports
goimports:
	@echo "Running goimports..."
	@find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./gen/*" -exec goimports -w {} \;

.PHONY: gofmt
gofmt:
	@echo "Running gofmt..."
	@find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./gen/*" -exec gofmt -w {} \;

.PHONY: format
format: goimports gofmt
	@echo "Formatting complete."

.PHONY: tests
tests:
	rm -f coverage.out
	go test ./... -count=1 -cover --coverprofile=coverage.out
