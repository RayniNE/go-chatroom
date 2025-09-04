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