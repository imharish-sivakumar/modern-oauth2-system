WORK_DIR = $(shell pwd)
PROJECT := token-management-service
REVISION := latest
BUILD_VENDOR := git config --global url."https://$(GITHUB_USERNAME):$(GITHUB_ACCESS_TOKEN)@github.com".insteadOf "https://github.com" && \
                go env -w GOPRIVATE=github.com/imharish-sivakumar && go mod vendor && chmod -R +w vendor

install_deps:
	docker compose -f infrastructure/build.yaml --project-name $(PROJECT) \
	run --rm build-env /bin/sh -c "apk update && apk add git && $(BUILD_VENDOR)"
build: install_deps
	docker compose -f infrastructure/build.yaml --project-name $(PROJECT) \
	run --rm build-env /bin/sh -c "go build -mod=vendor -ldflags '-s -w -X main.BuildRevision=$(CI_COMMIT_SHA) -X main.BuildVersion=$(CI_COMMIT_TAG) -X main.BuildTime=$(date --utc +%Y-%m-%dT%H:%M:%SZ)' -o ./bin/$(EXEC_NAME)"
dockerize: build
	docker compose -f docker-compose.app.yml build --no-cache
start:
	docker compose -f docker-compose.app.yml up -d
db-init:
	docker exec -it cisauth-db /bin/sh /sql/create_user.sh
stop:
	docker compose -f docker-compose.app.yml --project-name $(PROJECT) down -v
migration-up:
	migrate -path db/migrations -database "postgres://$(DB_SERVICE_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB)?sslmode=disable" -verbose up
migration-down:
	migrate -path db/migrations -database "postgres://$(DB_SERVICE_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB)?sslmode=disable" -verbose down
format:
	go fmt ./...
vet: install_deps
	docker compose -f infrastructure/build.yml --project-name $(PROJECT) \
	run --rm build-env /bin/sh -c "go vet -mod=vendor ./..."
hadolint:
	docker run --rm -i hadolint/hadolint:latest hadolint --ignore DL3007 --ignore DL3008 --ignore SC2016 - < infrastructure/Dockerfile