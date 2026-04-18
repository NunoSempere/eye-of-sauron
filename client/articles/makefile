include .env

build:
	go build -o eye src/*.go

run:
	go run src/*.go

trim:
	psql $(DATABASE_URL) -c "DELETE FROM sources WHERE processed = TRUE AND created_at < NOW() - INTERVAL '2 weeks';"
	psql $(DATABASE_URL) -c "SELECT * FROM sources WHERE processed = TRUE AND created_at < NOW() - INTERVAL '2 weeks';"

deps:
	go mod tidy
	go mod download
	go mod vendor

# RANDOM_CLUSTERS=true make
## SELECT COUNT(*) FROM your_table_name
# WHERE created_at >= NOW() - INTERVAL '8 HOUR';

