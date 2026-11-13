.PHONY: all build test clean web dev run demo report

BINARY=bin/llmmeter

all: build

web:
	cd web && npm install && npm run build

build: web
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/llmmeter

test:
	go test -v ./...

demo: build
	./$(BINARY) demo --db ./finops.db

run: build
	./$(BINARY) serve --port 8080 --db ./finops.db

report: build
	./$(BINARY) report --db ./finops.db

clean:
	rm -rf bin/ $(BINARY) finops.db*
