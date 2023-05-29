# Manifests for ingress controller deployment

This directory serves as a reference for deploying NanoProxy as an ingress controller to Kubernetes. These manifests are generally not to be used 'as-is' and should be modified as required.

A better way to deploy the NanoProxy ingress controller is using Helm.

- `deploy.yaml` - Runs both the `proxy` and `ingress-ctrl` images in a single pod, they share a config volume
- `ingress-class.yaml` - Creates an ingress class for the controller 
- `service-account.yaml` - Creates a `Role`, `ClusterRoleBinding` and `ServiceAccount` to support the controller
- `service.yaml` - Exposes the proxy using a *LoadBalancer* type service
  