.PHONY: build test vet lint install clean

BINARY=vibe-harness

build:
	go build -o $(BINARY) ./cmd/vibe-harness

test:
	go test ./...

vet:
	go vet ./...

lint: vet

install:
	go install ./cmd/vibe-harness

clean:
	rm -f $(BINARY)