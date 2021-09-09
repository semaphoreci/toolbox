test.setup:
	docker-compose build
	docker-compose run app go get ./...

test:
	docker-compose run --rm app gotestsum --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1

test.watch:
	docker-compose run --rm app gotestsum --watch --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1

build.mac:
	CGO_ENABLED=0 go build -o bin/cache

build.linux:
	CGO_ENABLED=0 GOOS=linux go build -o bin/cache
