.PHONY: dev server web test test-server test-web docker-up docker-down docs-public

dev:
	docker compose up --build

server:
	cd apps/server && go run ./cmd/eldermere

web:
	cd apps/web && npm run dev -- --host 0.0.0.0

test: test-server test-web

test-server:
	cd apps/server && go test ./...

test-web:
	cd apps/web && npm run check

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docs-public:
	npx docsify-cli serve docs/public
