# syntax=docker/dockerfile:1
FROM golang:1-alpine as builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o /phantom

FROM golang:1-alpine
COPY docker-entrypoint.sh /bin/docker-entrypoint.sh
COPY --from=builder /phantom /phantom
CMD [ "/bin/docker-entrypoint.sh"]
