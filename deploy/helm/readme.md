# Helm Chart

The chart is in the `./nanonproxy` folder, and provides an easy way to deploy NanoProxy into Kubernetes as an ingress controller.

It will create the following:

- *Deployment* for NanoProxy pod(s)
- *Service* to get traffic into the ingress controller & proxy
- *ClusterRole*, *ClusterRoleBinding* & *ServiceAccount* to support the ingress controller
- *IngressClass*

## Get Started

Github can be used as a remote Helm repo for installing the chart directly, without a need to clone the code repo.

```bash
helm repo add nanoproxy 'https://raw.githubusercontent.com/benc-uk/nanoproxy/main/deploy/helm'
helm repo update nanoproxy
```

```bash
helm install myingress nanoproxy/nanoproxy
```

## Full Chart Reference

Details on all the available chart values which can be set and passed in, are [contained in the chart readme](./nanoproxy/)
