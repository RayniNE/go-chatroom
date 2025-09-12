include $(PWD)/.env

run:
	@go run cmd/chatroom/main.go
install_migration:
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migration_up:
	@migrate -path migrations -database "$(DATABASE_URL)" up
migration_down:
	@migrate -path migrations -database "$(DATABASE_URL)" down
migration_drop:
	@migrate -path migrations -database "$(DATABASE_URL)" drop -f
run_test:
	@go test -count=1 -cover ./...
run_test_verbose:
	@go test -count=1 -v ./...
build_image:
	docker build -t chatroom-go .
start_image:
	docker run --env-file .env -p 8080:8080 -t chatroom-go