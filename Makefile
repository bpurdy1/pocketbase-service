LIBSQL_DOCKER_COMPOSE           = docker/libSQL/docker-compose.yml
LOCAL_SQLITE_FILE               = ./db/local_db/data.db
LOCAL_LOAD_DATA_DB_FILE         = ./db/turso/data.db


.PHONY: run build docker-dev libsql-dev libsql-down libsql-logs clean

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

docker-dev:
	docker-compose -f docker-compose.yml up -d



libsql-up:
	docker-compose -f $(LIBSQL_DOCKER_COMPOSE) up -d

libsql-down:
	docker-compose -f  $(LIBSQL_DOCKER_COMPOSE) down -v
libsql-clean:
	rm -rf ./docker/db/libsql/

load-data:
	@if [ -f $(LOCAL_LOAD_DATA_DB_FILE) ]; then \
		sqlite3 $(LOCAL_LOAD_DATA_DB_FILE) .dump > /tmp/pb_dump.sql && \
		turso db shell http://localhost:8080 < /tmp/pb_dump.sql && \
		rm -f $(LOCAL_LOAD_DATA_DB_FILE) $(LOCAL_LOAD_DATA_DB_FILE)-*; \
	else \
		echo "No local data.db found — skipping load"; \
	fi

down:
	docker-compose down
up:
	docker-compose up -d



clean:
	rm -rf bin/
	rm -rf db/
