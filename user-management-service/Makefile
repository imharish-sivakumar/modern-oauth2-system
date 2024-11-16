WORK_DIR = $(shell pwd)

PROJECT := user-management-service
REVISION := latest

BUILD_VENDOR := git config --global url."https://$(GITHUB_USERNAME):$(GITHUB_ACCESS_TOKEN)@github.com".insteadOf "https://github.com" && \
                go env -w GOPRIVATE=github.com/imharish-sivakumar && go mod vendor && chmod -R +w vendor

install_deps:
	docker compose -f infrastructure/build.yaml --project-name $(PROJECT) \
	run --rm build-env /bin/sh -c "apk update && apk add git && $(BUILD_VENDOR)"

build: install_deps
	docker compose -f infrastructure/build.yaml --project-name $(PROJECT) \
	run --rm build-env /bin/sh -c "go build -mod=vendor -ldflags '-s -w -X main.BuildRevision=$(CI_COMMIT_SHA) -X main.BuildVersion=$(CI_COMMIT_TAG) -X main.BuildTime=$(date --utc +%Y-%m-%dT%H:%M:%SZ)' -o ./bin/$(EXEC_NAME)"

db-init:
	psql -c 'CREATE DATABASE "user-management-service-go"' -U $(user)
migrationup:
	migrate -path db/migrations -database "postgres://$(user):$(password)@$(host):$(port)/user-management-service-go?sslmode=disable" -verbose up
migrationdown:
	migrate -path db/migrations -database "postgres://$(user):$(password)@$(host):$(port)/user-management-service-go?sslmode=disable" -verbose down