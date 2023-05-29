# Manifests for proxy deployment

This directory serves as a reference for deploying NanoProxy as a standalone proxy to Kubernetes. These manifests are generally not to be used 'as-is' and should be modified as required.

- `config.yaml` - The config file for the proxy. Modify this with your real config 😁
- `service.yaml` - Exposes the proxy using a *LoadBalancer* type service
- `deploy.yaml` - Runs `ghcr.io/benc-uk/nanoproxy-proxy:latest` as a deployment with a single pod
  