DSN=postgresql://postgres@localhost:5432?sslmode=disable
DB_CONTAINER_NAME=some-postgres
MIGRATIONS_DIR=db/migrations

init:
	go mod tidy

dev:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

mkdb:
	docker run --name "$(DB_CONTAINER_NAME)" --network host -e POSTGRES_PASSWORD=mysecretpassword -d postgres

execdb:
	docker exec -it "$(DB_CONTAINER_NAME)" psql -U postgres

rmdb:
	docker rm "$(DB_CONTAINER_NAME)" -f

migrateup:
	@docker start "$(DB_CONTAINER_NAME)"
	migrate -path "$(MIGRATIONS_DIR)" -database "$(DSN)" -verbose up

migratedown:
	@docker start "$(DB_CONTAINER_NAME)"
	migrate -path "$(MIGRATIONS_DIR)" -database "$(DSN)" -verbose down

sqlc:
	sqlc generate

proto:
	@rm -f pb/*.go
	protoc \
		--proto_path=proto \
		--go_out=pb \
		--go_opt=paths=source_relative \
		--go-grpc_out=pb \
		--go-grpc_opt=paths=source_relative \
	proto/*.proto

.PHONY: init dev mkdb execdb rmdb migrateup migratedown sqlc proto
