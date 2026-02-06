.PHONY: run build docker-dev libsql-dev libsql-down libsql-logs clean

run:
	go run cmd/server/main.go

run-turso:
	go run cmd/turso-server/main.go

build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/turso-server cmd/turso-server/main.go

docker-dev:
	docker-compose -f docker-compose.yml up -d

libsql-dev:
	docker-compose -f libSQL/docker-compose.yml up -d

down:
	docker-compose down 
up: 
	docker-compose up -d 

libsql-clean:
	docker-compose -f libSQL/docker-compose.yml down -v

clean:
	rm -rf bin/
	rm -rf db/
