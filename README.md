# bekind
Personal tool that sets up a KIND cluster to my personal specifications

Defaults to:

* Binds to ports 80/443 on host
* Installs NGINX Ingress controller
* Installs latest version of Argo CD
* "Multi Node" setup for KIND

# Config

You can customize the setup by providing a Specific Config (under `~/.bekind/config.yaml` or by providing `--config` to a YAML file)

For example:

* `domain`: Domain to use for any ingresses this tool will autocreate (assuming wildcard DNS)
* `kindImageVersion`: The KIND Node image to use (You can find a list [on dockerhub](https://hub.docker.com/r/kindest/node/tags))
* `kindConfig`: A custom [kind config](https://kind.sigs.k8s.io/docs/user/configuration/). It's "garbage in/garbage out" currently
* `helmCharts`: Different Helm Charts to install on startup. "garbage in/garbage out"

```yaml
domain: "7f000001.nip.io"
kindImageVersion: "kindest/node:v1.26.0"
helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-rollouts"
    release: "argo-rollouts"
    namespace: "argo-rollouts"
    args: 'installCRDs=true,controller.image.pullPolicy=IfNotPresent'
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
