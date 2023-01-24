SDB_CONTAINER_NAME=server-postgres
SDB_PORT=15432
CDB_CONTAINER_NAME=client-postgres
CDB_PORT=25432
MIGRATIONS_DIR=db/migrations
PASS=mysecretpassword
SDSN=postgresql://postgres:$(PASS)@localhost:$(SDB_PORT)?sslmode=disable

init:
	go mod tidy

dev:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

mkdb:
	docker run --name $(SDB_CONTAINER_NAME) -p $(SDB_PORT):5432 -e POSTGRES_PASSWORD=$(PASS) -d postgres
	docker run --name $(CDB_CONTAINER_NAME) -p $(CDB_PORT):5432 -e POSTGRES_PASSWORD=$(PASS) -d postgres

es:
	docker exec -it $(SDB_CONTAINER_NAME) psql -U postgres

ec:
	docker exec -it $(CDB_CONTAINER_NAME) psql -U postgres

rmdb:
	docker rm $(SDB_CONTAINER_NAME) -f
	docker rm $(CDB_CONTAINER_NAME) -f

migrateup:
	@docker start $(SDB_CONTAINER_NAME)
	migrate -path $(MIGRATIONS_DIR) -database "$(SDSN)" -verbose up

migratedown:
	@docker start $(SDB_CONTAINER_NAME)
	migrate -path $(MIGRATIONS_DIR) -database "$(SDSN)" -verbose down

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

.PHONY: init dev mkdb es ec rmdb migrateup migratedown sqlc proto
