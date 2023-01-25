PASS=mysecretpassword

SDB_CONTAINER_NAME=server-postgres
SDB_PORT=15432
SDSN=postgresql://postgres:$(PASS)@localhost:$(SDB_PORT)?sslmode=disable

CDB_CONTAINER_NAME=client-postgres
CDB_PORT=25432
CDSN=postgresql://postgres:$(PASS)@localhost:$(CDB_PORT)?sslmode=disable

MIGRATIONS_DIR=db/migrations

init:
	go mod tidy

dev:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

mkdb:
	docker run --name $(SDB_CONTAINER_NAME) -p $(SDB_PORT):5432 -e POSTGRES_PASSWORD=$(PASS) -d postgres
	docker run --name $(CDB_CONTAINER_NAME) -p $(CDB_PORT):5432 -e POSTGRES_PASSWORD=$(PASS) -d postgres
	@sleep 2

es:
	docker exec -it $(SDB_CONTAINER_NAME) psql -U postgres

ec:
	docker exec -it $(CDB_CONTAINER_NAME) psql -U postgres

rmdb:
	docker rm $(SDB_CONTAINER_NAME) -f
	docker rm $(CDB_CONTAINER_NAME) -f

refreshdb: rmdb mkdb migrateup

migrateup:
	@docker start $(SDB_CONTAINER_NAME)
	migrate -path $(MIGRATIONS_DIR) -database "$(SDSN)" -verbose up
	migrate -path $(MIGRATIONS_DIR) -database "$(CDSN)" -verbose up

migratedown:
	@docker start $(SDB_CONTAINER_NAME)
	migrate -path $(MIGRATIONS_DIR) -database "$(SDSN)" -verbose down
	migrate -path $(MIGRATIONS_DIR) -database "$(CDSN)" -verbose down

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

.PHONY: init dev mkdb es ec rmdb refreshdb migrateup migratedown sqlc proto
