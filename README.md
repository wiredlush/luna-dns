<p align="center">
  <a href="https://github.com/wiredlush/luna-dns">
     <picture>
    <img width="150" src="assets/logo.svg"/>
    </picture>
  </a>
  <br />
  <b>luna-dns</b>
  <br /> <br />
 <img src="https://img.shields.io/github/v/release/wiredlush/luna-dns" alt="Release" />
  <a href="https://github.com/wiredlush/luna-dns/actions/workflows/build.yml" /><img src="https://github.com/wiredlush/luna-dns/actions/workflows/build.yml/badge.svg" alt="Build"></a>
  <a href="https://github.com/wiredlush/luna-dns/actions/workflows/test.yml" /><img src="https://github.com/wiredlush/luna-dns/actions/workflows/test.yml/badge.svg" alt="Test"></a>
</p>
<h4 align="center">Straightforward DNS forwarder with cache, custom hosts and blocklists support</h3>

---

## Deployment

### Standalone Executable

<p align="justify">
In order to run luna-dns as a standalone executable, you can build it from source code or download a pre-built binary from the latest release.
</p>

**Build from source:**

```bash
git clone https://github.com/wiredlush/luna-dns.git
cd luna-dns
make
```

**Run executable:**

```bash
./luna-dns <path-to-config-file>
```

### Docker

<p align="justify">
You can deploy an instance of luna-dns by using Docker:
</p>

```bash
docker run -d --name=luna-dns \
  -p 5355:5355/udp \
  -v /path/to/config.yml:/etc/luna-dns/config.yml \
  --restart unless-stopped \
  wiredlush/luna-dns:latest
```

<p align="justify">
The luna-dns image will always look for a configuration file in /etc/luna-dns/config.yml
</p>

## Configuration

<p align="justify">
Luna DNS is configured via a YAML configuration file. A sample configuration is provided in the root of this repository (config.yml).
</p>

```yml
addr: "0.0.0.0:5355" # the address luna-dns will bind to
network: "udp" # luna-dns server protocol (udp or tcp supported)
cache_ttl: 14400 # after how long the cached data is cleared (expressed in seconds)

# if a valid file path is set the logs will be written to that file
# leave empty to disable log file
log_file: "test.log"

# remote dns servers to forward requests to (if not matching custom hosts)
dns:
  - addr: "8.8.8.8:53" # remote dns addr string
    network: "tcp" # remote dns protocol (udp or tcp supported)
  - addr: "8.8.4.4:53"
    network: "udp"

# custom hosts entries
hosts:
  - host: google.com # custom host domain or pattern (wildcards supported)
    ip: 127.0.1.1 # custom host ip
  - host: "test.com"
    ip: 127.0.0.1
  - host: "*.test.com" # wildcard pattern example
    ip: 127.0.0.1

# luna-dns supports blocklists both from local files or remote URI
# Blocklists must contain only one domain name per line.
# Every blocked record resolves to 0.0.0.0
# ex.
# google.com
# test.com
# ...
blocklists:
  - http://test.test/test.txt
  - file://folder/test.txt
  - file:///root/test.txt

# blocklists update time (in minute)
blocklists_update: 720 # 12 hours (default)
```
