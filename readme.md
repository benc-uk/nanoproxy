# NanoProxy

This is simple HTTP reverse proxy written in Go and based largely on [httputil.ReverseProxy](https://pkg.go.dev/net/http/httputil#ReverseProxy) in the Go standard library

Features:
- Host and path based routing, with prefix and exact matching modes.
- Strip path support, removes the matching path before sending on the request.
- Preserves the host header for the upstream requests, like [any good reverse proxy should](https://learn.microsoft.com/en-us/azure/architecture/best-practices/host-name-preservation).
- The headers `X-Forwarded-For`, `X-Forwarded-Host`, `X-Forwarded-Proto` are set on the upstream request.

It can run standalone, as a container and also operate as a Kubernetes ingress controller, utilizing the standard `Ingress` resource and API

## Config

Config is done with YAML and consists of arrays of two main objects, `upstreams` and `rules`. Upstreams are the target servers you want to send requests onto. 
Rules are routing rules for matching requests and assigning them to one of the upstreams

By default the file `config.yaml` is loaded, a different name can be set with `-config` argument when starting the proxy.

### Upstream
```yaml
name:   Name
host:   Hostname or IP
port:   Port number, if omitted it defaults to 80 or 443
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

## Routing and matching logic