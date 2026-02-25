.PHONY: start generate-jwt test test-unit test-integration lint

start:
	docker compose up

generate-jwt:
	go run ./cmd/token_generator

test-unit:
	go test ./...

test-integration:
	docker compose -p avito-tests -f docker-compose.test.yaml up --build --abort-on-container-exit --exit-code-from tests

test:
	@$(MAKE) test-unit && $(MAKE) test-integration

lint:
	golangci-lint run
