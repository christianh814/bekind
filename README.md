# bekind
Personal tool that sets up a KIND cluster to my personal specifications

# Config

Specific Config:

`domain`: Domain to use for any ingresses this tool will autocreate (assuming wildcard DNS)
`kindConfig`: A custom [kind config](https://kind.sigs.k8s.io/docs/user/configuration/). It's "garbage in/garbage out" currently

```yaml
domain: "7f000001.nip.io"
kindConfig: |
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  networking:
    disableDefaultCNI: True
    podSubnet: "10.254.0.0/16"
    serviceSubnet: "172.30.0.0/16"
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 80
      hostPort: 80
      listenAddress: 0.0.0.0
    - containerPort: 443
      hostPort: 443
      listenAddress: 0.0.0.0
```
