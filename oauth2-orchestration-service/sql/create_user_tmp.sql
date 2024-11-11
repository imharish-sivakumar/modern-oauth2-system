DROP USER IF EXISTS hydra;
CREATE USER hydra WITH PASSWORD 'secret';
DROP DATABASE IF EXISTS hydra;
CREATE DATABASE hydra WITH OWNER hydra;
GRANT ALL PRIVILEGES ON DATABASE hydra TO hydra;

DROP USER IF EXISTS "user-service";
CREATE USER "user-service" WITH PASSWORD 'secret';
DROP DATABASE IF EXISTS "user-management-service";
CREATE DATABASE "user-management-service" WITH OWNER "user-service";
GRANT ALL PRIVILEGES ON DATABASE "user-management-service" TO "user-service";
