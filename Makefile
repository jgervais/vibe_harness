.PHONY: build test vet lint harness install clean

BINARY=vibe-harness

all: build lint

build:
	go build -o $(BINARY) ./cmd/vibe-harness

test:
	go test ./...

vet:
	go vet ./...

harness: build
	./$(BINARY) --format json .

lint: vet harness

install:
	go install ./cmd/vibe-harness

clean:
	rm -f $(BINARY)
