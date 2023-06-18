# Samples

This directory has a few sample services and ingress definitions you can deploy to Kubernetes for testing

- **httpbin** - This Deploys [httpbin](http://httpbin.org/) as a Pod in the cluster and creates a NanoProxy ingress to
  route to it. Requests beginning with `/httpbin` will route to this
- **httpbin-ext** - This creates an `ExternalName` service to point at http://httpbin.org/ and also creates a NanoProxy
  ingress to route to it. Requests beginning with `/ext` will route to this
