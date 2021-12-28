BINARY_NAME=IntraProxy

GOBASE=$(bash pwd)
GOPATH=$(GOBASE)
GOBIN=$(GOBASE)/bin

build:
	@echo "  >  Build for Linux"
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o bin/${BINARY_NAME} ./cmd/

run:
	@echo "  >  Running"
	go run ./cmd

clean:
	@echo "  >  Delete binary..."
	go clean
	rm ./bin/${BINARY_NAME}