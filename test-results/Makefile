PHONY: run

run:
	go run main.go $(arg)

regen.golden:
	go test ./pkg/parsers -update-golden

test.setup:
	docker compose build
	docker compose run cli go get ./...

lint:
	docker compose run --rm cli revive -formatter json -config lint.toml ./... > revive.json
	docker compose run --rm cli staticcheck -f json ./... > staticcheck.json

test:
	docker compose run --rm cli gotestsum --format short-verbose --junitfile junit-report.xml --packages="./..."

test.watch:
	gotestsum --watch ./...

test.cover:
	go install github.com/jandelgado/gcov2lcov@latest
	go install github.com/securego/gosec/v2/cmd/gosec@v2.19.0
	go test -coverprofile=c.out ./...
	gcov2lcov -infile=c.out -outfile=coverage.lcov
	rm c.out
	scripts/lcov-to-md.sh

MONOREPO_TMP_DIR?=/tmp/monorepo
SECURITY_TOOLBOX_TMP_DIR?=$(MONOREPO_TMP_DIR)/security-toolbox
SECURITY_TOOLBOX_BRANCH?=main
LANGUAGE=go
APP_DIRECTORY=/app

check.prepare:
	rm -rf $(MONOREPO_TMP_DIR)
	git clone --depth 1 --filter=blob:none --sparse https://github.com/semaphoreio/semaphore $(MONOREPO_TMP_DIR) && \
		cd $(MONOREPO_TMP_DIR) && \
		git config core.sparseCheckout true && \
		git sparse-checkout init --cone && \
		git sparse-checkout set security-toolbox && \
		git checkout main && cd -

check.static: check.prepare
	docker run -it \
		-v $$(pwd):$(APP_DIRECTORY) \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:3 \
		bash -c 'cd $(APP_DIRECTORY) && $(SECURITY_TOOLBOX_TMP_DIR)/code --language $(LANGUAGE) -d'

check.deps: check.prepare
	docker run -it \
		-v $$(pwd):$(APP_DIRECTORY) \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		-e TRIVY_DB_REPOSITORY -e TRIVY_JAVA_DB_REPOSITORY \
		registry.semaphoreci.com/ruby:3 \
		bash -c 'cd $(APP_DIRECTORY) && $(SECURITY_TOOLBOX_TMP_DIR)/dependencies --language $(LANGUAGE) -d'

build.darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=$(ARCH) go build -o bin/darwin/$(ARCH)/test-results main.go

build.linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o bin/linux/$(ARCH)/test-results main.go

build.windows:
	CGO_ENABLED=0 GOOS=windows go build -o bin/windows/test-results.exe main.go
