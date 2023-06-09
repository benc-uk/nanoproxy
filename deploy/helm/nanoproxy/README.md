# nanoproxy

![Version: 0.0.3](https://img.shields.io/badge/Version-0.0.3-informational?style=flat-square)
![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)
![AppVersion: 0.0.3](https://img.shields.io/badge/AppVersion-0.0.3-informational?style=flat-square)

NanoProxy ingress controller

## Values

| Key                        | Type   | Default                       | Description                                                                                                                            |
| -------------------------- | ------ | ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| affinity                   | object | `{}`                          | Affinity for pod assignment                                                                                                            |
| debug                      | string | `""`                          | Turn on debug logging in the proxy                                                                                                     |
| fullnameOverride           | string | `""`                          | Override the fullname used when creating resources                                                                                     |
| image.prefix               | string | `"ghcr.io/benc-uk/nanoproxy"` | Prefix for the image repository, '-proxy' and '-controller' will be appended                                                           |
| image.pullPolicy           | string | `"Always"`                    | Image pull policy                                                                                                                      |
| image.tag                  | string | `""`                          | Overrides the image tag whose default is the chart appVersion.                                                                         |
| imagePullSecrets           | list   | `[]`                          | Set the imagePullSecrets value to enable pulling images from private registry                                                          |
| ingressClass.create        | bool   | `true`                        | Create an IngressClass resource                                                                                                        |
| ingressClass.name          | string | `"nanoproxy"`                 | Name of the IngressClass resource                                                                                                      |
| nameOverride               | string | `""`                          | Override the release name used when creating resources                                                                                 |
| nodeSelector               | object | `{}`                          | Node selector for pod assignment                                                                                                       |
| podAnnotations             | object | `{}`                          | Pod annotations                                                                                                                        |
| podSecurityContext         | object | `{}`                          | Security context for the pods                                                                                                          |
| replicaCount               | int    | `1`                           | Replica count                                                                                                                          |
| resources.limits.cpu       | string | `"200m"`                      | CPU resource limits                                                                                                                    |
| resources.limits.memory    | string | `"128Mi"`                     | Memory resource limits                                                                                                                 |
| securityContext            | object | `{}`                          | Security context for the containers                                                                                                    |
| service.loadBalancerIP     | string | `nil`                         | Use an existing IP address for the service                                                                                             |
| service.port               | int    | `80`                          | Port to expose on the service, change to 443 if using TLS                                                                              |
| service.type               | string | `"LoadBalancer"`              | Type of service to create                                                                                                              |
| serviceAccount.annotations | object | `{}`                          | Annotations to add to the service account                                                                                              |
| serviceAccount.create      | bool   | `true`                        | Specifies whether a service account should be created This will also create ClusterRole and ClusterRoleBinding for the service account |
| serviceAccount.name        | string | `""`                          | The name of service account to use. If not set and create is true, name is generated                                                   |
| tls.enabled                | bool   | `false`                       | Enable TLS on the proxy                                                                                                                |
| tls.secretName             | string | `""`                          | TLS secret name, must be set if enabled is true                                                                                        |
| tolerations                | list   | `[]`                          | Tolerations for pod assignment                                                                                                         |
