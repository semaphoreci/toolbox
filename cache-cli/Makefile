test.setup:
	go get ./...
	go get gotest.tools/gotestsum

test:
	gotestsum --junitfile /tmp/junit.xml

test.watch:
	gotestsum --watch ./...

build:
	CGO_ENABLED=0 go build -o bin/cache
