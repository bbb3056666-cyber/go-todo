APP := todo-server

.PHONY: run test lint build

run:
	go run ./cmd/server

test:
	go test ./... -race -cover

lint:
	go vet ./...

build:
	go build -o bin/$(APP) ./cmd/server

clean:
	rm -rf bin

fmt:
	gofmt -l -w .

check: fmt lint test