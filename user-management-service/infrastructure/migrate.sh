#!/usr/bin/env sh
SECRET_JSON=$(aws secretsmanager get-secret-value --secret-id $SECRET_KEY --query 'SecretString' --output text)
ENV_VARS=$(echo $SECRET_JSON | jq -r 'to_entries[] | "\(.key)=\(.value)"')

# Set the environment variables
eval $ENV_VARS
export POSTGRES_DB_PASSWORD=$POSTGRES_DB_PASSWORD

if [ $? -eq 0 ]; then
    /home/gola/flyway -url=jdbc:postgresql://"${POSTGRES_DB_HOST}":"${POSTGRES_DB_PORT}"/"${POSTGRES_DB_NAME}" -schemas="${POSTGRES_DB_NAME}" -user="${POSTGRES_DB_USER}" -password="${POSTGRES_DB_PASSWORD}" -connectRetries=60 -mixed=true migrate
else
    echo "Unable to substitute credentials"
    exit 1
fi
