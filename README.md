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
* Docker (Podman is still [considered experemental](https://kind.sigs.k8s.io/docs/design/principles/#target-cri-functionality) by KIND)


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
* `postInstallManifests`: List of YAML files to apply to the KIND cluster after setup. There is no checks done and any errors are from the K8S API. Currently only YAML files are supported. It's "garbage in/garbage out".
* `postInstallActions`: Actions to perform on Kubernetes resources after installation. See [Post Install Actions](#post-install-actions) for more info.

```yaml
domain: "7f000001.nip.io"
kindImageVersion: "kindest/node:v1.34.0"
helmCharts:
  - url: "https://kubernetes.github.io/ingress-nginx"
    repo: "ingress-nginx"
    chart: "ingress-nginx"
    release: "nginx-ingress"
    namespace: "ingress-controller"
    wait: true
    valuesObject:
      controller:
        extraArgs:
          enable-ssl-passthrough: ""
        hostNetwork: true
        ingressClassResource:
          default: true
        nodeSelector:
          nginx: ingresshost
        service:
          externalTrafficPolicy: ""
          type: ClusterIP
        tolerations:
        - operator: Exists
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: true
    valuesObject:
      configs:
        secret:
          argocdServerAdminPassword: $2a$10$pKM9yRpR2G5X8c3.M.lgs.v5xBBzEyiJnH5vrWYGkO3JNr5HTW8yq
        cm:
          kustomize.buildOptions: "--enable-helm"
          resource.customizations.health.argoproj.io_Application: |
            hs = {}
            hs.status = "Progressing"
            hs.message = ""
            if obj.status ~= nil then
              if obj.status.health ~= nil then
                hs.status = obj.status.health.status
                if obj.status.health.message ~= nil then
                  hs.message = obj.status.health.message
                end
              end
            end
            return hs
          resource.customizations.health.networking.k8s.io_Ingress: |
            hs = {}
            hs.status = "Healthy"
            hs.message = "Probably just fine"
            return hs
      global:
        domain: argocd.7f000001.nip.io
      server:
        ingress:
          annotations:
            '"nginx.ingress.kubernetes.io/force-ssl-redirect"': true
            '"nginx.ingress.kubernetes.io/ssl-passthrough"': true
          enabled: true
          hostname: argocd.7f000001.nip.io
          ingressClassName: nginx
          tls: true
kindConfig: |
  kind: Cluster
  name: argocd-ingress
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
    - quay.io/christianh814/simple-go:latest
    - quay.io/ablock/gitops-helm-quarkus:latest
    - christianh814/gobg:latest
postInstallManifests:
  - "file:///home/chernand/workspace/argocd/bunch-o-apps/helm-app-example.yaml"
  - "file:///home/chernand/workspace/argocd/bunch-o-apps/gobg.yaml"
  - "file:///home/chernand/workspace/argocd/bunch-o-apps/simple-go.yaml"
postInstallActions:
  - action: restart
    kind: Deployment
    name: argocd-server
    namespace: argocd
  - action: restart
    kind: Deployment
    namespace: argocd
    labelSelector:
      app.kubernetes.io/name: argocd-applicationset-controller
  - action: delete
    kind: Pod
    name: old-pod
    namespace: default
  - action: delete
    kind: Pod
    namespace: kube-system
    labelSelector:
      app: cleanup
```

# Helm Chart Config

The following are valid configurations for the `helmCharts` section:

* `url`: The URL of the Helm repo (*REQUIRED*). Can be OCI repo with `oci://`
* `repo`: What to name the repo, interally (*REQUIRED*). It's the `<reponame>` from `helm repo add <reponame> <url>`. (ignored when using OCI)
* `chart`: What chart to install from the Helm repo (*REQUIRED*). (Ignored when using OCI)
* `release`: What to call the release when it's installed (*REQUIRED*).
* `namespace`: The namespace to install the release to, it'll create the namespace if it's not already there (*REQUIRED*).
* `version`: The version of the Helm chart to install (*Optional*)
* `wait`: Wait for the release to be installed before returning (*Optional*); default is `false`.
* `valuesObject`: A YAML object to use as the values file (*Optional*).

# Loading Docker Images

The following are valid configurations for the `loadDockerImages` section:

> :rotating_light: **NOTE** Only `docker` is supported see [KIND upstream issue](https://github.com/kubernetes-sigs/kind/issues/2038)

* `pullImages`: To perform a pull of the image before lodaing (opional and defaults to `true` if not supplied). This is a "global" setting (you're either pulling them all or none)
* `images`: List of images to do the pull.

# Post Install Actions

The `postInstallActions` section allows you to perform actions on Kubernetes resources after the cluster setup is complete. This runs after Helm charts are installed and manifests are applied.

Currently supported actions:
- `restart`: Performs a rollout restart on the specified resources (equivalent to `kubectl rollout restart`)
- `delete`: Deletes the specified resources (equivalent to `kubectl delete`)

## Configuration Options

* `action`: The action to perform (*REQUIRED*). Supported values: `restart`, `delete`.
* `kind`: The kind of Kubernetes resource (*REQUIRED*). 
  - For `restart`: `Deployment`, `StatefulSet`, `DaemonSet`
  - For `delete`: `Pod`
* `namespace`: The namespace where the resource(s) exist (*Optional*). Defaults to `default` if not specified.
* `name`: The name of the specific resource (*Optional* - either `name` or `labelSelector` must be provided).
* `labelSelector`: A map of labels to select multiple resources (*Optional* - either `name` or `labelSelector` must be provided).
* `group`: The API group (*Optional*). Defaults to `core` for Pod, `apps` for Deployment/StatefulSet/DaemonSet.
* `version`: The API version (*Optional*). Defaults to `v1`.

## Restart by Name

To restart a specific resource by name:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    name: argocd-server
    namespace: argocd
```

This is equivalent to running:
```bash
kubectl rollout restart deployment/argocd-server -n argocd
```

## Restart by Label Selector

To restart all resources matching a label selector:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    namespace: argocd
    labelSelector:
      app.kubernetes.io/name: argocd-applicationset-controller
```

This is equivalent to running:
```bash
kubectl rollout restart deployment -n argocd -l app.kubernetes.io/name=argocd-applicationset-controller
```

You can use multiple labels in the selector:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    namespace: my-namespace
    labelSelector:
      app: myapp
      environment: production
```

**Note:** If both `name` and `labelSelector` are provided, `labelSelector` takes precedence and `name` is ignored.

## Delete by Name

To delete a specific Pod by name:

```yaml
postInstallActions:
  - action: delete
    kind: Pod
    name: foo
    namespace: bar
```

This is equivalent to running:
```bash
kubectl delete pod/foo -n bar
```

## Delete by Label Selector

To delete all Pods matching a label selector:

```yaml
postInstallActions:
  - action: delete
    kind: Pod
    namespace: bazz
    labelSelector:
      app.kubernetes.io/instance: bizz
```

This is equivalent to running:
```bash
kubectl delete pod -n bazz -l app.kubernetes.io/instance=bizz
```
