HTMX_VERSION=2.0.6


.PHONY: download-htmx
## download-htmx: Downloads HTMX minified js file
download-htmx:
	curl -o web/static/htmx.min.js https://cdn.jsdelivr.net/npm/htmx.org@${HTMX_VERSION}/dist/htmx.min.js 

.PHONY: build
## build: Builds the Go program
build:
	CGO_ENABLED=0 \
	go build -o monitor .

.PHONY: dev
dev: build-css
	templ generate --watch --cmd="go generate" &\
	templ generate --watch --cmd="go run main.go server"

.PHONY: create-migration
## create: Creates golang-migrate migration files
create-migration:
	 migrate create -ext sql -dir ./db/migrations $(file_name)

.PHONY: css
## css: Watch and compile Tailwind CSS for development
css:
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/styles.css --watch

.PHONY: build-css
## build-css: Build and minify Tailwind CSS for production
build-css:
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/styles.css --minify

.PHONY: create-secrets
## create-secrets: create necessary podman secrets
create-secrets:
	@echo -n "${NOSTRICH_WATCH_MONITOR_PRIVATE_KEY}" | podman secret create nostrich-watch-monitor-private-key -
	@echo -n "${NOSTRICH_WATCH_DB_PASSWORD}" | podman secret create nostrich-watch-db-password -

.PHONY: setup-services
## setup-services: copy the content of the quadlet directory to ~/.config/containers/systemd/nostrich-watch/ and reload the system files
setup-services:
	mkdir -p ~/.config/containers/systemd/
	cp quadlet/* ~/.config/containers/systemd/
	systemctl --user daemon-reload

.PHONY: setup-nostrich-watch-db
## setup-nostrich-watch-db: setup the database by enabling and starting the database service. 
setup-nostrich-watch-db:
	systemctl --user start nostrich-watch-db  

.PHONY: setup-config
## setup-config: copy config files to the right location
setup-config:
	mkdir -p ~/.config/nostrich-watch/prometheus/targets
	cp config.toml ~/.config/nostrich-watch/
	cp prometheus/prometheus.yml ~/.config/nostrich-watch/prometheus/
	cp prometheus/targets/services.yml ~/.config/nostrich-watch/prometheus/targets/
	cp prometheus/targets/asynqmon.yml ~/.config/nostrich-watch/prometheus/targets/

.PHONY: start-services
## start-services: start all services after migrations
start-services:
	systemctl --user start nostrich-watch-cache nostr-rs-relay nostrich-watch-job-scheduler nostrich-watch-dashboard nostrich-watch-worker nostrich-watch-asynqmon

.PHONY: stop-services
## stop-services: stop all services
stop-services:
	systemctl --user stop nostrich-watch-cache nostr-rs-relay nostrich-watch-job-scheduler nostrich-watch-dashboard nostrich-watch-worker nostrich-watch-asynqmon

.PHONY: list-services
## list-services: list all services
list-services:
	systemctl --user list-unit-files | grep nostrich
	
.PHONY: enable-linger
## enable-linger: enable the linger for our user to start the containers without the user being logged in
enable-linger:
	loginctl enable-linger "${USER}"

.PHONY: help
## help: Prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
