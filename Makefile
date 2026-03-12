.PHONY: all web frontend clean

all:
	CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o build/luna-dns ./cmd/luna-dns

web: frontend
	CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -tags web -o build/luna-dns-web ./cmd/luna-dns

frontend:
	cd web && npm install && npm run build

clean:
	rm -rf build pkg/web/static
