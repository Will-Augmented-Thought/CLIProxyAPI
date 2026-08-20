FROM docker.io/library/golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG VERSION=dev
ARG BUILD_GIT_SHA
ARG BUILD_TIMESTAMP

RUN test -n "${BUILD_GIT_SHA}" \
    && test -n "${BUILD_TIMESTAMP}" \
    && CGO_ENABLED=1 GOOS=linux go build -buildvcs=false -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Commit=${BUILD_GIT_SHA}' -X 'main.BuildDate=${BUILD_TIMESTAMP}'" -o ./CLIProxyAPI ./cmd/server/

FROM docker.io/library/debian:bookworm

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=builder ./app/CLIProxyAPI /CLIProxyAPI/CLIProxyAPI

COPY config.example.yaml /CLIProxyAPI/config.example.yaml

WORKDIR /CLIProxyAPI

EXPOSE 8317

ENV TZ=Asia/Shanghai

CMD ["./CLIProxyAPI"]
