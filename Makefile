APP_NAME := taper

.PHONY: run build test fmt lint tidy clean

run:
	go run .

build:
	go build -o $(APP_NAME) .

test:
	go test ./...

fmt:
	gofmt -w .

lint:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(APP_NAME)
