.DEFAULT: build

.PHONY: build
build: clean test
	mkdir -p bin
	go build -o bin/index_reader cmd/main.go

.PHONY: clean
clean:
	rm -f bin/*

.PHONY: test
test:
	go test ./...
