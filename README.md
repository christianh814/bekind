# BeKind

Installs a K8S cluster using [KIND](https://github.com/kubernetes-sigs/kind), and does a number of post deployment steps.

Bekind will:

* Install a KIND cluster based on the supplied config
* KIND cluster can be modified to deploy a specific K8S version
* Installs any Supplied Helm Charts
* Loads images into the KIND cluster

# Installation

Prerequisites:

* go version `1.20` or newer
* Docker (Podman is still [considered experemental](https://github.com/kubernetes-sigs/kind/pull/1302) by KIND)


Install with:

```shell
go install github.com/christianh814/bekind@latest
```

Then move into your `$PATH` (example showing `/usr/local/bin`)

```shell
sudo mv $GOBIN/bekind /usr/local/bin/bekind
sudo chmod +x /usr/local/bin/bekind
```

# Config

You can customize the setup by providing a Specific Config (under `~/.bekind/config.yaml` or by providing `--config` to a YAML file)

For example:

* `domain`: Domain to use for any ingresses this tool will autocreate, assuming wildcard DNS (currently unused/ignored)
* `kindImageVersion`: The KIND Node image to use (You can find a list [on dockerhub](https://hub.docker.com/r/kindest/node/tags)). You can also supply your own public image or a local image.
* `kindConfig`: A custom [kind config](https://kind.sigs.k8s.io/docs/user/configuration/). It's "garbage in/garbage out".
* `helmCharts`: Different Helm Charts to install on startup. "garbage in/garbage out". See [Helm Chart Config](#helm-chart-config) for more info.
* `loadDockerImages`: List of images to load onto the nodes. See the [Loading Docker Images](#loading-docker-images) section below for more info.
* `postInstallManifests`: List of YAML files to apply to the KIND cluster after setup. This is the last step to run in the process. There is no checks done and any errors are from the K8S API. Currently only YAML files are supported. It's "garbage in/garbage out".

```yaml
domain: "7f000001.nip.io"
kindImageVersion: "kindest/node:v1.29.1"
helmCharts:
  - url: "https://kubernetes.github.io/ingress-nginx"
    repo: "ingress-nginx"
    chart: "ingress-nginx"
    release: "nginx-ingress"
    namespace: "ingress-controller"
    args:
      - name: 'controller.hostNetwork'
        value: "true"
      - name: 'controller.nodeSelector.nginx'
        value: "ingresshost"
      - name: 'controller.service.type'
        value: "ClusterIP"
      - name: 'controller.tolerations[0].operator'
        value: "Exists" - name: 'controller.service.externalTrafficPolicy'
        value: ""
      - name: 'controller.extraArgs.enable-ssl-passthrough'
        value: ""
    wait: true
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    args:
      - name: 'server.ingress.enabled'
        value: "true"
      - name: 'server.ingress.hosts[0]'
        value: "argocd.7f000001.nip.io"
      - name: 'server.ingress.ingressClassName'
        value: "nginx"
      - name: 'server.ingress.https'
        value: "true"
      - name: 'server.ingress.annotations."nginx\.ingress\.kubernetes\.io/ssl-passthrough"'
        value: "true"
      - name: 'server.ingress.annotations."nginx\.ingress\.kubernetes\.io/force-ssl-redirect"'
        value: "true"
    wait: true
  - url: "https://redhat-developer.github.io/redhat-helm-charts"
    repo: "redhat-helm-charts"
    chart: "quarkus"
    release: "myapp"
    namespace: "demo"
    version: "0.0.3"
    args:
      - name: 'build.enabled'
        value: "false"
      - name: 'deploy.route.enabled'
        value: "false"
      - name: 'image.name' 
        value: "quay.io/ablock/gitops-helm-quarkus"
    wait: true
  - url: "oci://ghcr.io/akuity/kargo-charts/kargo"
    repo: "kargo"
    chart: "kargo"
    release: "kargo"
    namespace: "kargo"
    args: 
      - name: 'api.adminAccount.password'
        value: "admin"
      - name: 'controller.logLevel'
        value: "DEBUG"
      - name: 'api.adminAccount.tokenTTL'
        value: "24h"
      - name: 'api.adminAccount.tokenSigningKey'
        value: "secret"
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
  pullImages: true
  images:
    - gcr.io/kuar-demo/kuard-amd64:blue
postInstallManifests:
  - 'file:///path/to/local/k8s/file.yaml'
  - 'https://raw.githubusercontent.com/christianh814/gitops-examples/main/gobg/gobg.yaml'
```

# Helm Chart Config

The following are valid configurations for the `helmCharts` section:

* `url`: The URL of the Helm repo (*REQUIRED*). Can be OCI repo with `oci://`
* `repo`: What to name the repo, interally (*REQUIRED*). It's the `<reponame>` from `helm repo add <reponame> <url>`. (ignored when using OCI)
* `chart`: What chart to install from the Helm repo (*REQUIRED*). (Ignored when using OCI)
* `release`: What to call the release when it's installed (*REQUIRED*).
* `namespace`: The namespace to install the release to, it'll create the namespace if it's not already there (*REQUIRED*).
* `version`: The version of the Helm chart to install (*Optional*)
* `args`: The parameter of the `--set` command to change the values in a comma separated format. This is a list of key value pairs using `name` for the key and `value` for the value. (*Optional*)
* `wait`: Wait for the release to be installed before returning (*Optional*); default is `false`.

# Loading Docker Images

The following are valid configurations for the `loadDockerImages` section:

> :rotating_light: **NOTE** Only `docker` is supported see [KIND upstream issue](https://github.com/kubernetes-sigs/kind/pull/3109)

* `pullImages`: To perform a pull of the image before lodaing (opional and defaults to `true` if not supplied). This is a "global" setting (you're either pulling them all or none)
* `images`: List of images to do the pull.
