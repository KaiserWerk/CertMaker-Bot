FROM golang:1.27 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG BUILD_DATE
RUN BUILD_DATE="${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" && CGO_ENABLED=0 go build -ldflags "-X main.Version=${VERSION} -X main.VersionDate=${BUILD_DATE}" -o certmaker-bot ./cmd/certmaker-bot
FROM alpine:3.20 AS final-stage
RUN apk add --no-cache bash ca-certificates
WORKDIR /app
COPY --from=build-stage /app/certmaker-bot .
ENTRYPOINT ["/app/certmaker-bot"]