.PHONY: env run web build test lint fmt generate docker

env:
	test -f core/.env || cp core/.env.example core/.env

run:
	cd core && go run ./cmd/app

web:
	cd web && npm run dev

build:
	cd web && npm run build
	cd core && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/fluffy-giggle ./cmd/app

test:
	cd core && go test -race ./...
	cd web && npm test

lint:
	cd core && golangci-lint run ./...
	cd web && npm run lint && npm run typecheck && npm run format:check

fmt:
	cd core && golangci-lint fmt
	cd web && npm run format

generate:
	cd core && go tool sqlc generate && go generate ./internal/...

docker:
	docker build -t fluffy-giggle:local .
