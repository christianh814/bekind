# bekind
Personal tool that sets up a KIND cluster to my personal specifications

# Config

Specific Config:

* `domain`: Domain to use for any ingresses this tool will autocreate (assuming wildcard DNS)
* `kindConfig`: A custom [kind config](https://kind.sigs.k8s.io/docs/user/configuration/). It's "garbage in/garbage out" currently
* `helmCharts`: Different Helm Charts to install on startup. "garbage in/garbage out"

```yaml
domain: "7f000001.nip.io"
helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-rollouts"
    release: "argo-rollouts"
    namespace: "argo-rollouts"
    args: 'installCRDs=true'
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
