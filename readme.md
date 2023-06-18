# NanoProxy

<img src="docs/logo.png" align="left" width="200px"/>

NanoProxy is a simple HTTP reverse proxy & Kubernetes ingress controller written in Go and based largely on
[httputil.ReverseProxy](https://pkg.go.dev/net/http/httputil#ReverseProxy) in the Go standard library. It was designed
for traffic routing (like an API gateway) and less for load balancing.

This was developed as a learning exercise only! If you want an ingress controller for your production Kubernetes cluster
you [should look elsewhere](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/).

<br clear="left"/>

Features:

- Host and path based routing, with prefix and exact matching modes.
- Can run as a Kubernetes ingress controller, using the core `Ingress` resource and utilizes the sidecar pattern.
- Strip path support, removes the matching path before sending on the request.
- Preserves the host header for the upstream requests, like
  [any good reverse proxy should](https://learn.microsoft.com/en-us/azure/architecture/best-practices/host-name-preservation).
- The headers `X-Forwarded-For`, `X-Forwarded-Host`, `X-Forwarded-Proto` are set on the upstream request.
- HTTPS support with TLS termination.

### Container Images

Prebuilt containers are published on GitHub Container Registry

[![](https://ghcr-badge.egpl.dev/benc-uk/nanoproxy-proxy/tags?ignore=none&label=proxy)](https://github.com/benc-uk/nanoproxy/pkgs/container/nanoproxy-proxy)
[![](https://ghcr-badge.egpl.dev/benc-uk/nanoproxy-controller/tags?ignore=none&label=controller)](https://github.com/benc-uk/nanoproxy/pkgs/container/nanoproxy-controller)

### Project Status

![](https://img.shields.io/github/last-commit/benc-uk/nanoproxy)
![](https://img.shields.io/github/release/benc-uk/nanoproxy)
![](https://img.shields.io/github/actions/workflow/status/benc-uk/nanoproxy/ci-build.yml?branch=main)

## ☸️ Deploying to Kubernetes

Using the NanoProxy Helm chart is the recommended way to install as an ingress controller into your cluster

- [Docs for the ingress controller Helm chart](deploy/helm/readme.md)

If you really don't want to use Helm for some reason, basic manifests are also provided to deploy as:

- [A Ingress controller](deploy/kubernetes/ingress-ctrl)
- [A standalone reverse proxy](deploy/kubernetes/proxy)

## 🐋 Running the proxy as container

You can simply run:

```bash
docker run -p 8080:8080 ghcr.io/benc-uk/nanoproxy-proxy:latest
```

But this isn't very helpful, as you will be running with an empty configuration! To mount a local folder containing a
config file locally, try the following:

```bash
docker run -p 8080:8080 \
-v $PATH_TO_CONF:/conf \
-e CONF_FILE=/conf/config.yaml \
ghcr.io/benc-uk/nanoproxy-proxy:latest
```

## 🎯 Ingress Controller

The ingress controller (or just controller) works by listening to the Kubernetes API and watching for `Ingress`
resources. It then reconciles each `Ingress` using an in memory cache (simply a map keyed on namespace & name) and the
following logic:

1. Detect if the action is a deletion, if _Ingress_ matches one in the cache and has been removed from Kubernetes, if so
   remote it from the cache.
2. Check the _Ingress_ has an `ingressClassName` in the spec, matching by name an _IngressClass_, and this
   _IngressClass_ resource matches our controller ID `benc-uk/nanoproxy`. Exit if there is no match.
3. Add the _Ingress_ to the cache or update existing one based on key.
4. Build NanoProxy configuration from cache, mapping fields from the _Ingress_ spec into upstreams and rules (see
   [proxy config below](#🛠️-proxy-config)).
5. Write configuration file.

The controller needs to run as a sidecar beside the proxy, this is achieved by running both containers in the same pod,
and using a shared volume so the config file written by the controller is picked up by the proxy. This is best explained
with a diagram:

![Diagram of NanoProxy running as an Ingress Controller](./docs/diagram.drawio.png)

The controller was created using the Operator SDK, roughly following
[this guide](https://kubernetes.io/blog/2021/06/21/writing-a-controller-for-pod-labels/)

### Ingress Controller - Annotations

The following annotations are supported:

- `nanoproxy/backend-protocol` - Specify 'http' or 'https', default is 'http'
- `nanoproxy/strip-path` - Strip the path, 'true' or 'false', see proxy config below. Note this will apply to all
  rules/routes under this _Ingress_, create multiple _Ingresses_ if you need a mix. Default is 'false'

## 🛠️ Proxy Config

NanoProxy configuration is done with YAML and consists of arrays of two main objects, `upstreams` and `rules`. Upstreams
are the target servers you want to send requests onto. Rules are routing rules for matching requests and assigning them
to one of the upstreams.

The proxy process watches the config file for changes and will reload the configuration if the file is updated.

> Note. When running as an ingress controller you do not supply a config file, as it is completely managed by the
> controller.

By default the file `./config.yaml` local to current directory of the binary, a different filename & path can be set
with `-config` or `-c` argument when starting the proxy.

### Upstream

```yaml
name: Name
host: Hostname or IP
port: Port number, defaults to 80 or 443 when scheme is https
scheme: Scheme 'http' or 'https', if omitted defaults to 'http'
```

### Rule

```yaml
upstream: Name of the upstream to send traffic to
path: URL path in request to match against
host: Host in request to match against. If omitted, will match all hosts
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

## ⚙️ Environmental Variables

| Env Var           | Description                                                                                                                                    | Default |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `CONF_FILE`       | Used by both the proxy and the controller, path of the config file used.                                                                       | _None_  |
| `TIMEOUT`         | Connection and HTTP timeout in seconds. Proxy only.                                                                                            | 5       |
| `PORT`            | Port the proxy will listen and accept traffic on.                                                                                              | 8080    |
| `DEBUG`           | For extra logging and output from the proxy, set to non-blank value (e.g. "1"). Also enables the special config endpoint (see below).          | _None_  |
| `CERT_PATH`       | Set to a directory where `cert.pem` and `key.pem` reside, this will enable TLS and HTTPS on the proxy server.                                  | _None_  |
| `TLS_SKIP_VERIFY` | Used when calling a HTTPS upstream, if this var is set to anything (e.g. "1") this will skip the normal TLS cert validation for all upstreams. | _None_  |

## 🤖 Notes on proxy

The proxy exposes two routes of it's own:

- `/.nanoproxy/health` Returns HTTP 200 OK. Used for health checks, and probes
- `/.nanoproxy/config` Dumps the in memory config, this endpoint is only enabled when DEBUG is set

The proxy applies the following logic to incoming requests to decide how to route them:

- Get hostname from incoming request
  - Loop over all the `rules`
  - If the rule has a `host` set, match it with the hostname
  - OR if the rule has an empty `host` field
    - Match the request path to the rule `path`, matching can be `prefix` or `exact`
    - If match is made this `rule` is selected and no further rules are checked
      - Get the matching named `upstream` referenced by the `rule`
      - Pass HTTP request to the reverse proxy for that `upstream`

## 🧑‍💻 Developer Guide

It's advised to use the published container image rather than trying to run from source, but if you wish to try running
the code yourself, here's some getting started details

### Pre-requisites

- Go 1.20+
- Bash and make
- Docker or other container runtime engine

The makefile should help you carry out most tasks. Linters and supporting tools are installed into a local `.tools`
directory. Run `make install-tools` to download and set these up.

Then use `make run-proxy` or `make run-ctrl` to run either or both locally.

```
$ make
build                🔨 Build binary into ./bin/ directory
clean                🧹 Clean up, remove dev data and files
helm-package         🔠 Package Helm chart and update index
help                 💬 This help message :)
images               📦 Build container images
install-tools        🔮 Install dev tools into project bin directory
lint-fix             📝 Lint & format, attempts to fix errors & modify code
lint                 🔍 Lint & format check only, sets exit code on error for CI
print-env            🚿 Print all env vars for debugging
push                 📤 Push container images
release              🚀 Release a new version on GitHub
run-ctrl             👟 Run controller locally with hot-reload
run-proxy            👟 Run proxy locally with hot-reload
test                 🧪 Run all unit tests
```

### Repo Index

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
