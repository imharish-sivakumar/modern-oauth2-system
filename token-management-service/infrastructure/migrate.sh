#!/usr/bin/env sh
git config --global url."https://${GITHUB_USERNAME}:${GITHUB_ACCESS_TOKEN}@${CI_SERVER_HOST}".insteadOf "https://github.com"
go env -w GONOPROXY="github.com/affordmed/*"
go env -w GONOSUMDB="github.com/affordmed/*"
go env -w GOPRIVATE="github.com/affordmed/*"
go install github.com/affordmed/azure-services-client@latest
DB_PASSWORD=$(azure-services-client -key DevUMSDBPassword)
export DB_PASSWORD=$DB_PASSWORD
if [ $? -eq 0 ]; then
    migrate -path /database -database postgres://"${DB_USER}":"${DB_PASSWORD}"@"${DB_HOST}":"${DB_PORT}"/"${DB_NAME}"?sslmode=disable -verbose up
else
    echo "Unable to retrieve password"
fi
