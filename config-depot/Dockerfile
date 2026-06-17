FROM golang:1.25-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/config-depot ./cmd/config-depot
RUN mkdir -p /out/data/config

FROM scratch

WORKDIR /app

ENV CONFIG_DEPOT_ADDR=:8080
ENV CONFIG_DEPOT_DATA_DIR=/data

EXPOSE 8080
VOLUME ["/data"]

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build --chown=65532:65532 /out/data /data
COPY --from=build /out/config-depot /usr/local/bin/config-depot

USER 65532:65532
ENTRYPOINT ["config-depot"]
