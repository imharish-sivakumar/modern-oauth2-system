# oauth2-orchestration-service

## Getting started

To run an oauth2 server locally, use this docker compose scripts to run inside docker.

---

## Pre requisites
* Install docker
* Install docker compose
* export a variables in your ~/.zshrc or ~/.bashrc
    ```
      export ENVIRONMENT=local
    ```

#### Note: If you are running in windows, then try using WSL2 for running the scripts.

----

### Create required env files
* Create file named ```docker-compose-local.env``` and add values for below keys
    ```
    POSTGRES_USER=<postgres-sys-username>
    POSTGRES_PASSWORD=<postgres-password>
    PGPASSWORD=<pg-password>
    ```

## Run Server
### Running make start target will complete below operations 
* *Creates a docker network for cisauth*
* *Creates a postgres db.*
* *Creates required service users (RBAC).*
* *Runs redis cache.*
* *Starts ORY/Hydra(oauth2 server)*
* *Creates oauth2 clients*
* *All the above services will be running under project name cisauth*
### Start Server
```
make start
```
---
## Stop Server
* Will destroy everything under the project name cisauth. [containers, volumes, network].
```
make stop
```

---
#### Use the oauth2-scrapper to generate a session id & access token for a user

### Issues in using windows chocolaty make
1. Even if we use chocolaty make, still environmental placeholders inside the docker compose file won't work.
2. If you prefer to run in normal PS/CMD replace the ${} with respective value and copy and run individual command inside make targets.

#### Justifications:
1. Why we need to create multiple users in the sql/create_user.sql file ?
   1. to achieve database per service one shouldn't use common root user(postgres) as root user have admin privileges. if that user is exposed to the intruder not only the service but all database schemas will be affected.
   2. achieving RBAC from the local to prod.
2. Avoiding running separate postgres application/redis application required for our services, instead running required technologies under docker hood.