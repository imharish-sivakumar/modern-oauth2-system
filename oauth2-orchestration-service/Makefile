
PROJECT := cisauth

ifeq ($(findstring local,$(ENVIRONMENT)),local)
UMS_CALLBACK := http://localhost:3000/user-service/v1/login/accept
else
UMS_CALLBACK := https://cisauth.org/user-service/v1/login/accept
endif

create_network:
	docker network create cisauth-network

docker_login:
	@docker login -u $(ARTIFACTORY_USER) -p $(ARTIFACTORY_PASSWORD)

create_user:
	docker exec -it cisauth-db bash -c "/sql/create_user.sh"

create-db:
	docker volume prune -f && \
	docker compose -f docker-compose.db.yaml --project-name $(PROJECT) up -d cisauth-db && \
	sleep 20

create-hydra:
	docker compose -f docker-compose.db.yaml --project-name $(PROJECT) up --build -d hydra-migrate hydra admin-hydra-migrate admin-hydra && \
	sleep 30

run-hydra: create-hydra
	docker compose --project-name $(PROJECT) exec hydra hydra create client --endpoint http://localhost:4445 --name ui-web-client --response-type code,id_token,token --grant-type implicit,refresh_token,authorization_code,client_credentials --scope openid,offline,offline_access,profile,email,address,phone,api --redirect-uri $(UMS_CALLBACK) --token-endpoint-auth-method 'client_secret_basic' --secret $$(date +'%Y%m%d%H%M%S') && \
	docker compose --project-name $(PROJECT) exec hydra hydra create client --endpoint http://localhost:4445 --name forgot-credential-client --secret $$(date +'%Y%m%d%H%M%S') --grant-type client_credentials --scope api && \
	docker compose --project-name $(PROJECT) exec hydra hydra create client --endpoint http://localhost:4445 --name verify-user-client --secret $$(date +'%Y%m%d%H%M%S') --grant-type client_credentials --scope api

run_cache:
	docker compose -f docker-compose.db.yaml --project-name $(PROJECT) up -d cisauth-cache

start: create_network create-db create_user run_cache run-hydra

stop:
	docker compose -f docker-compose.db.yaml --project-name $(PROJECT) down -v && \
	docker image prune -f || true && \
	docker network rm cisauth-network
