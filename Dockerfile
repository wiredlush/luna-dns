FROM golang:1.23 AS go-builder
WORKDIR /luna-dns
COPY . .
RUN make

FROM alpine:3.17 AS luna-dns
WORKDIR /etc/luna-dns
COPY ./config.yml .
WORKDIR /usr/bin
COPY --from=go-builder ./luna-dns/build/luna-dns .
ENTRYPOINT ["/usr/bin/luna-dns", "/etc/luna-dns/config.yml"]
