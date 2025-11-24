lint:
	go mod tidy
	command -v goimports || go install golang.org/x/tools/cmd/goimports@latest
	@if [ -n "$$(go env GOBIN)" ]; then \
		$$(go env GOBIN)/goimports -w .; \
	else \
		$$(go env GOPATH)/bin/goimports -w .; \
	fi
	go vet ./...
	gofmt -w .
	go run mvdan.cc/unparam@latest
	go run golang.org/x/tools/cmd/deadcode@latest -test ./...
