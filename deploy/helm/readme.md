# Helm Chart

The chart is in the `./nanonproxy` folder, and provides an easy way to deploy NanoProxy into Kubernetes as an ingress
controller.

It will create the following:

- _Deployment_ for NanoProxy pod(s)
- _Service_ to get traffic into the ingress controller & proxy
- _ClusterRole_, _ClusterRoleBinding_ & _ServiceAccount_ to support the ingress controller
- _IngressClass_

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

Details on all the available chart values which can be set and passed in, are
[contained in the chart readme](./nanoproxy/)
