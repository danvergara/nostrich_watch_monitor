.PHONY: build
## build: Builds the Go program
build:
	CGO_ENABLED=0 \
	go build -o monitor .

.PHONY: create-migration
## create: Creates golang-migrate migration files
create-migration:
	 migrate create -ext sql -dir ./db/migrations $(file_name)

.PHONY: help
## help: Prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
