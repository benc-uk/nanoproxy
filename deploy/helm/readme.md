# Helm Chart

This Helm chart provides an easy way to deploy NanoProxy into Kubernetes as an ingress controller.

Github can be used as a remote Helm repo for installing the chart directly

```bash
helm repo add nanoproxy 'https://raw.githubusercontent.com/benc-uk/nanoproxy/main/deploy/helm'
helm repo update nanoproxy

helm install myrelease nanoproxy/nanoproxy
```

## Full Chart Reference

Details on all the available chart values which can be set and passed in, are [contained in the chart readme](./nanoproxy/)
