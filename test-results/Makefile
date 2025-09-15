PHONY: run

run:
	go run main.go $(arg)

regen:
	go run main.go compile --no-compress priv/parsers/junit_generic/in.xml priv/parsers/junit_generic/out.json
	go run main.go compile --no-compress priv/parsers/junit_rspec/in.xml priv/parsers/junit_rspec/out.json
	go run main.go compile --no-compress priv/parsers/junit_exunit/in.xml priv/parsers/junit_exunit/out.json
	go run main.go compile --no-compress priv/parsers/junit_golang/in.xml priv/parsers/junit_golang/out.json
	go run main.go compile --no-compress -p phpunit priv/parsers/junit_phpunit/in.xml priv/parsers/junit_phpunit/out.json
	go run main.go compile --no-compress -p embedded priv/parsers/junit_embedded/in.xml priv/parsers/junit_embedded/out.json
	go run main.go compile --no-compress -p go:staticcheck priv/parsers/go_staticcheck/in.json priv/parsers/go_staticcheck/out.json
	go run main.go compile --no-compress priv/merging priv/merging/out.json

test.setup:
	docker compose build
	docker compose run cli go get ./...

lint:
	docker compose run --rm cli revive -formatter friendly -config lint.toml ./...

test:
	docker compose run --rm cli gotestsum --format short-verbose --junitfile junit-report.xml --packages="./..."

test.parsers:
	go test ./pkg/parsers -run TestParsersWithGoldenFiles

test.parsers.update:
	go test ./pkg/parsers -run TestParsersWithGoldenFiles -update-golden

update-golden: test.parsers.update
	@echo "Golden files updated successfully"

test.watch:
	gotestsum --watch ./...

test.cover:
	go install github.com/jandelgado/gcov2lcov@latest
	go install github.com/securego/gosec/v2/cmd/gosec@v2.19.0
	go test -coverprofile=c.out ./...
	gcov2lcov -infile=c.out -outfile=coverage.lcov
	rm c.out
	scripts/lcov-to-md.sh

SECURITY_TOOLBOX_BRANCH ?= master
SECURITY_TOOLBOX_TMP_DIR ?= /tmp/security-toolbox

check.prepare:
	rm -rf $(SECURITY_TOOLBOX_TMP_DIR)
	git clone git@github.com:renderedtext/security-toolbox.git $(SECURITY_TOOLBOX_TMP_DIR) && (cd $(SECURITY_TOOLBOX_TMP_DIR) && git checkout $(SECURITY_TOOLBOX_BRANCH) && cd -)

check.static: check.prepare
	docker run -it -v $$(pwd):/app \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:2.7 \
		bash -c 'cd /app && $(SECURITY_TOOLBOX_TMP_DIR)/code --language go -d'

check.deps: check.prepare
	docker run -it -v $$(pwd):/app \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		-e TRIVY_DB_REPOSITORY -e TRIVY_JAVA_DB_REPOSITORY \
		registry.semaphoreci.com/ruby:2.7 \
		bash -c 'cd /app && $(SECURITY_TOOLBOX_TMP_DIR)/dependencies --language go -d'

build.darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=$(ARCH) go build -o bin/darwin/$(ARCH)/test-results main.go

build.linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o bin/linux/$(ARCH)/test-results main.go

build.windows:
	CGO_ENABLED=0 GOOS=windows go build -o bin/windows/test-results.exe main.go
