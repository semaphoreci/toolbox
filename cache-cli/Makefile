.PHONY: test

test.setup:
	docker-compose build
	docker-compose run cli go get ./...
	docker-compose up -d

test:
	docker-compose run --rm cli gotestsum --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1

test.watch:
	docker-compose run --rm cli gotestsum --watch --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1

build:
	GOOS=$(OS) go build -o bin/$(OS)/cache main.go
