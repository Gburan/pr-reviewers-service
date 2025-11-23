LINTER = golangci-lint
LINTER_FLAGS = run

.DEFAULT_GOAL := lint

.PHONY: lint
lint:
	$(LINTER) $(LINTER_FLAGS)

.PHONY: lint-fix
lint-fix:
	$(LINTER) $(LINTER_FLAGS) --fix

.PHONY: install
install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang/mock/gomockgen@latest

.PHONY: t-clean
run-test-clean:
	go clean -testcache
	go test -p 1 ./...

.PHONY: gen
gen:
	go generate ./...

.PHONY: run-f
run-compose-f:
	go mod tidy
	go generate ./...
	docker-compose up -d

.PHONY: run
run-compose:
	docker-compose up -d

.PHONY: run-b
run-compose-b:
	docker-compose up -d --build

.PHONY: stress-test
stress-test:
	docker-compose up -d
	docker-compose up k6-load-test
