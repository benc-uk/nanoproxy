# NanoProxy

This is simple HTTP reverse proxy written in Go and based largely on [httputil.ReverseProxy](https://pkg.go.dev/net/http/httputil#ReverseProxy) in the Go standard library

Features:
- Host and path based routing, with prefix and exact matching modes.
- Can run as a Kubernetes ingress controller, utilizing the core `Ingress` resource.
- Strip path support, removes the matching path before sending on the request.
- Preserves the host header for the upstream requests, like [any good reverse proxy should](https://learn.microsoft.com/en-us/azure/architecture/best-practices/host-name-preservation).
- The headers `X-Forwarded-For`, `X-Forwarded-Host`, `X-Forwarded-Proto` are set on the upstream request.

## 📂 Repo Index

```text
📂
├── build         - Docker build files
├── deploy      
│   ├─ manifests  - Kubernetes manifests to deploy as ingress controller
│   └─ helm       - Helm chart to deploy as ingress controller
├── ingress-ctrl  - Source code of the ingress controller
├── pkg           - Shared packages between proxy and controller
├── proxy         - Source code of the reverse proxy
└── samples       - Samples and examples
```

## ☸️ Deploying to Kubernetes

Blah blah see `deploy/kubernetes`  
Blah blah Helm blah `deploy/helm`

## 🐋 Running the proxy as container

You can simply run:

```bash
docker run -p 8080:8080 ghcr.io/benc-uk/nanoproxy-proxy:latest
```

But this isn't very helpful! As you will be running with an empty configuration. 
To mount a local folder containing your config file locally, try the following:

```bash
docker run -p 8080:8080 \
-v $PATH_TO_CONF:/conf \
-e CONF_FILE=/conf/config.yaml \
ghcr.io/benc-uk/nanoproxy-proxy:latest
```

## 🧑‍💻 Developing Locally

### Pre-requisites  

- Go 1.20+
- Bash and make 
- Docker or other container runtime engine

The makefile should help you get started with this repo

```
$ make
help                 💬 This help message :)
install-tools        🔮 Install dev tools into project bin directory
lint                 🔍 Lint & format check only, sets exit code on error for CI
lint-fix             📝 Lint & format, attempts to fix errors & modify code
build                🔨 Build binary into ./bin/ directory
images               📦 Build container images
push                 📤 Push container images
run-proxy            🌐 Run proxy locally with hot-reload
run-ctrl             🤖 Run controller locally with hot-reload
test                 🧪 Run all unit tests
clean                🧹 Clean up, remove dev data and files
```

Run `make install-tools` then use `make run-proxy` or `make run-ctrl` to run either or both locally

## ⚙️ Environmental Vars

- `CONF_FILE`: This is used by both the proxy and the controller to set the path of the config file.
- `TIMEOUT`: Connection and HTTP timeout used by the proxy.
- `PORT`: Port the proxy will listen on.

## 🎯 Ingress Controller 

## 🛠️ Proxy Config

When running NanoProxy as a standalone reverse proxy, config is done with YAML and consists of arrays of two main objects, `upstreams` and `rules`. Upstreams are the target servers you want to send requests onto. 
Rules are routing rules for matching requests and assigning them to one of the upstreams

By default the file `config.yaml` is loaded, a different name can be set with `-config` argument when starting the proxy.

### Upstream

```yaml
name:   Name
host:   Hostname or IP
port:   Port number, defaults to 80 or 443 when scheme is https
scheme: Scheme 'http' or 'https', if omitted defaults to 'http'
```

### Rule

```yaml
upstream:  Name of the upstream to send traffic to
path:      URL path in request to match against
host:      Host in request to match against. If omitted, will match all hosts
matchMode: How to match the path, 'prefix' or 'exact', defaults to 'prefix'
stripPath: Remove the path before sending to upstream, true/false, defaults to false
```

Example config

```yaml
upstreams:
  - name: my-server-a
    host: some.hostname.here
    scheme: https
  - name: my-server-b
    host: backend.api.example
    port: 3000

rules:
  - upstream: my-server-b
    path: /api
    stripPath: true
  - upstream: my-server-a
    path: /
    host: proxy.example.net
```

## 🤖 Routing and matching logic

