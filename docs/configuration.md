---
layout: default
title: Configuration
nav_order: 3
description: "How to configure BeKind"
---

# Configuration
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

BeKind can be customized using a YAML configuration file. By default, BeKind looks for configuration at `~/.bekind/config.yaml`, but you can specify a custom location using the `--config` flag.

```bash
bekind start --config /path/to/custom/config.yaml
```

---

## Configuration File Structure

Here's a complete example showing all available configuration options:

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
        hostNetwork: true
kindConfig: |
  kind: Cluster
  name: my-cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  nodes:
  - role: control-plane
loadDockerImages:
  pullImages: true
  images:
    - gcr.io/kuar-demo/kuard-amd64:blue
postInstallManifests:
  - "file:///path/to/manifest.yaml"
postInstallActions:
  - action: restart
    kind: Deployment
    name: my-deployment
    namespace: default
```

---

## Configuration Options

### domain

**Type**: `string`  
**Optional**: Yes  
**Description**: Domain to use for any ingresses that BeKind might autocreate. Assumes wildcard DNS.

```yaml
domain: "7f000001.nip.io"
```

{: .note }
Currently unused/ignored in most workflows, but reserved for future features.

---

### kindImageVersion

**Type**: `string`  
**Optional**: Yes  
**Default**: Latest stable KIND node image  
**Description**: The KIND node image to use. This determines which Kubernetes version your cluster will run.

```yaml
kindImageVersion: "kindest/node:v1.34.0"
```

You can find available versions on [Docker Hub](https://hub.docker.com/r/kindest/node/tags). You can also supply your own public image or a local image.

---

### kindConfig

**Type**: `string` (multiline)  
**Optional**: Yes  
**Description**: A custom [KIND cluster configuration](https://kind.sigs.k8s.io/docs/user/configuration/). This is passed directly to KIND.

```yaml
kindConfig: |
  kind: Cluster
  name: my-cluster
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
          node-labels: "ingress=host"
    extraPortMappings:
    - containerPort: 80
      hostPort: 80
    - containerPort: 443
      hostPort: 443
```

{: .warning }
**"Garbage in/garbage out"** - BeKind passes this configuration directly to KIND without validation. Errors will come from KIND or the Kubernetes API.

---

### helmCharts

**Type**: `array`  
**Optional**: Yes  
**Description**: List of Helm charts to install after cluster creation.

See the [Helm Charts feature documentation]({% link features/helm-charts.md %}) for detailed information.

**Example**:

```yaml
helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: true
    version: "5.46.0"
    valuesObject:
      server:
        service:
          type: NodePort
```

---

### loadDockerImages

**Type**: `object`  
**Optional**: Yes  
**Description**: Configuration for loading Docker images into the KIND cluster.

See the [Loading Docker Images feature documentation]({% link features/loading-images.md %}) for detailed information.

**Example**:

```yaml
loadDockerImages:
  pullImages: true
  images:
    - gcr.io/kuar-demo/kuard-amd64:blue
    - quay.io/christianh814/simple-go:latest
```

---

### postInstallManifests

**Type**: `array`  
**Optional**: Yes  
**Description**: List of Kubernetes YAML manifest files to apply after cluster setup. Supports both local files (`file://`) and remote URLs (`http://` or `https://`).

See the [Post Install Manifests feature documentation]({% link features/post-install-manifests.md %}) for detailed information.

**Example**:

```yaml
postInstallManifests:
  - "file:///home/user/k8s/app.yaml"
  - "https://example.com/configs/service.yaml"
  - "file:///home/user/k8s/ingress.yaml"
```

---

### postInstallActions

**Type**: `array`  
**Optional**: Yes  
**Description**: Actions to perform on Kubernetes resources after installation.

See the [Post Install Actions feature documentation]({% link features/post-install-actions.md %}) for detailed information.

**Example**:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    name: argocd-server
    namespace: argocd
  - action: delete
    kind: Pod
    namespace: default
    labelSelector:
      app: cleanup
```

---

## Configuration Profiles

BeKind supports configuration profiles, which allow you to save and reuse different cluster configurations.

### Creating a Profile

Profiles are stored in `~/.bekind/profiles/<profile-name>/config.yaml`.

For example, to create an "argocd" profile:

```bash
mkdir -p ~/.bekind/profiles/argocd
nano ~/.bekind/profiles/argocd/config.yaml
```

You can also have multiple YAML configuration files in the same profile directory. All `.yaml` files in the profile directory will be executed when you run the profile.

### Using a Profile

Use the `run` command to execute a profile:

```bash
bekind run argocd
```

This will look for configuration files in `~/.bekind/profiles/argocd/` and execute each one.

### Custom Profile Directory

If your profiles are stored in a different location, use the `--profile-dir` flag:

```bash
bekind run myprofile --profile-dir /path/to/profiles
```

For example, if your config is at `/tmp/foo/config.yaml`:

```bash
bekind run foo --profile-dir /tmp
```

### Viewing Profile Configuration

To view a profile's configuration without running it:

```bash
bekind run argocd --view
```

### Default Configuration

If you don't want to use profiles, you can use `bekind start` with a config file:

```bash
bekind start --config ~/.bekind/config.yaml
```

Or place your config at `~/.bekind/config.yaml` and BeKind will use it by default.

---

## Examples

### Minimal Configuration

```yaml
kindImageVersion: "kindest/node:v1.34.0"
```

### Development Cluster with Ingress

```yaml
domain: "127.0.0.1.nip.io"
kindImageVersion: "kindest/node:v1.34.0"
kindConfig: |
  kind: Cluster
  name: dev
  apiVersion: kind.x-k8s.io/v1alpha4
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 80
      hostPort: 80
    - containerPort: 443
      hostPort: 443
helmCharts:
  - url: "https://kubernetes.github.io/ingress-nginx"
    repo: "ingress-nginx"
    chart: "ingress-nginx"
    release: "nginx-ingress"
    namespace: "ingress-nginx"
    wait: true
```

---

## Next Steps

Learn more about specific features:

- [Helm Charts]({% link features/helm-charts.md %})
- [Loading Docker Images]({% link features/loading-images.md %})
- [Post Install Manifests]({% link features/post-install-manifests.md %})
- [Post Install Actions]({% link features/post-install-actions.md %})
