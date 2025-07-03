run:
	@go run ./cmd/main.go

tidy:
	@go mod tidy

remove:
	@docker stop $$(docker ps -q) 2>/dev/null || true
	@docker rm -f $$(docker ps -aq) 2>/dev/null || true

compose:
	@docker compose up --build