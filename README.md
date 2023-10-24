# BeKind

Installs a K8S cluster using KIND, and does a number of post deployment steps.

Bekind will:

* Installs a KIND cluster based on the supplied config
* KIND cluster can be modified to deploy a specific K8S version
* Installs any Supplied Helm Charts (OCI registries not supported currently...PRs welcome!)
* Loads images into the KIND cluster (image MUST exist locally currently...again...PRs are welcome!)

# Config

You can customize the setup by providing a Specific Config (under `~/.bekind/config.yaml` or by providing `--config` to a YAML file)

For example:

* `domain`: Domain to use for any ingresses this tool will autocreate, assuming wildcard DNS (currently unused)
* `kindImageVersion`: The KIND Node image to use (You can find a list [on dockerhub](https://hub.docker.com/r/kindest/node/tags)). You can also supply your own public image or a local image.
* `kindConfig`: A custom [kind config](https://kind.sigs.k8s.io/docs/user/configuration/). It's "garbage in/garbage out".
* `helmCharts`: Different Helm Charts to install on startup. "garbage in/garbage out". See [Helm Chart Config](#helm-chart-config) for more info.
* `loadDockerImages`: List of images to load onto the nodes (**NOTE** images must exist locally). Only `docker` is supported (see [KIND upstream issue](https://github.com/kubernetes-sigs/kind/pull/3109))

```yaml
domain: "7f000001.nip.io"
kindImageVersion: "kindest/node:v1.28.0"
helmCharts:
  - url: "https://kubernetes.github.io/ingress-nginx"
    repo: "ingress-nginx"
    chart: "ingress-nginx"
    release: "nginx-ingress"
    namespace: "ingress-controller"
    args: 'controller.hostNetwork=true,controller.nodeSelector.nginx=ingresshost,controller.service.type=ClusterIP,controller.service.externalTrafficPolicy=,controller.extraArgs.enable-ssl-passthrough=,controller.tolerations[0].operator=Exists'
    wait: true
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    args: 'server.ingress.enabled=true,server.ingress.hosts[0]=argocd.7f000001.nip.io,server.ingress.ingressClassName="nginx",server.ingress.https=true,server.ingress.annotations."nginx\.ingress\.kubernetes\.io/ssl-passthrough"=true,server.ingress.annotations."nginx\.ingress\.kubernetes\.io/force-ssl-redirect"=true'
    wait: true
  - url: "https://redhat-developer.github.io/redhat-helm-charts"
    repo: "redhat-helm-charts"
    chart: "quarkus"
    release: "myapp"
    namespace: "demo"
    version: "0.0.3"
    args: 'build.enabled=false,deploy.route.enabled=false,image.name=quay.io/ablock/gitops-helm-quarkus'
    wait: true
kindConfig: |
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  networking:
    podSubnet: "10.254.0.0/16"
    serviceSubnet: "172.30.0.0/16"
  nodes:
  - role: control-plane
    kubeadmConfigPatches:
    - |
      kind: InitConfiguration
      nodeRegistration:
        kubeletExtraArgs:
          node-labels: "nginx=ingresshost"
    extraPortMappings:
    - containerPort: 80
      hostPort: 80
      listenAddress: 0.0.0.0
    - containerPort: 443
      hostPort: 443
      listenAddress: 0.0.0.0
loadDockerImages:
  - gcr.io/kuar-demo/kuard-amd64:blue
```
# Helm Chart Config

The following are valid configurations for the `helmCharts` section:

* `url`: The URL of the Helm repo (*REQUIRED*)
* `repo`: What to name the repo, interally (*REQUIRED*). It's the `<reponame>` from `helm repo add <reponame> <url>`.
* `chart`: What chart to install from the Helm repo (*REQUIRED*).
* `release`: What to call the release when it's installed (*REQUIRED*).
* `namespace`: The namespace to install the release to, it'll create the namespace if it's not already there (*REQUIRED*).
* `version`: The version of the Helm chart to install (*Optional*)
* `args`: The parameter of the `--set` command to change the values in a comma separated format. (*REQUIRED*)
* `wait`: Wait for the release to be installed before returning (*Optional*); default is `false`.
