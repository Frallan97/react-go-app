FROM golang:1.20 AS build
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:3.18

# install certificates, curl and migrate
RUN apk add --no-cache ca-certificates curl \
 && curl -L \
    https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz \
  | tar xz -C /usr/local/bin migrate \
 && chmod +x /usr/local/bin/migrate

WORKDIR /app
COPY --from=build /go/src/app/server /usr/local/bin/server
COPY migrations /migrations

# DB URL will come from the env
ENV DB_URL=postgres://postgres:postgres@db:5432/postgres?sslmode=disable

ENTRYPOINT [ "/bin/sh", "-c", "migrate -path /migrations -database \"$DB_URL\" up && /usr/local/bin/server" ] 