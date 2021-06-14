.PHONY: install
install: build
	go install ./cmd/tfconvert/.

.PHONY: build
build: generate test 
	go build ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	go test -v ./...