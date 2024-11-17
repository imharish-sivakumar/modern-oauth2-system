FROM alpine:3.16

ARG ARCH
RUN apk add git

RUN apk add --no-cache curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-${ARCH}.tar.gz | tar xz
RUN chmod +x migrate
RUN mv migrate /usr/local/bin/

# Install AWS CLI
RUN apk add --no-cache python3 py3-pip
RUN pip3 install awscli

WORKDIR /src
COPY ./db/migrations /database
COPY infrastructure/migrate.sh /src

ENTRYPOINT ["/src/migrate.sh"]
