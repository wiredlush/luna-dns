ARG BUILD_MODE=cli

FROM node:20-alpine AS node-builder
WORKDIR /luna-dns/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ .
RUN npm run build

FROM golang:1.24 AS go-builder
ARG BUILD_MODE
WORKDIR /luna-dns
COPY . .
COPY --from=node-builder /luna-dns/pkg/web/static/ ./pkg/web/static/
RUN if [ "$BUILD_MODE" = "web" ]; then \
        CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -tags web -o build/luna-dns ./cmd/luna-dns; \
    else \
        make; \
    fi

FROM alpine:3.17 AS luna-dns
WORKDIR /etc/luna-dns
COPY ./config.yml .
WORKDIR /usr/bin
COPY --from=go-builder ./luna-dns/build/luna-dns .
ENTRYPOINT ["/usr/bin/luna-dns", "/etc/luna-dns/config.yml"]
