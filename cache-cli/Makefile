.PHONY: test

test.setup:
	docker-compose build
	docker-compose run cli go get ./...

test:
	docker-compose run --rm cli gotestsum --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1

test.watch:
	docker-compose run --rm cli gotestsum --watch --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1

build:
	docker-compose run --rm --no-deps cli bash -c 'rm -rf bin && GOOS=$(OS) go build -o bin/cache main.go'
