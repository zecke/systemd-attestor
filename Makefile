utils = github.com/goreleaser/goreleaser

build:
	go build -i

utils: $(utils)

$(utils): noop
	go get $@

test:
	go test -race

release:
	goreleaser || true

clean:
	go clean

noop:

.PHONY: all build vendor utils test clean
