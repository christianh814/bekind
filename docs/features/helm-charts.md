---
layout: default
title: Helm Charts
parent: Features
nav_order: 1
description: "Install Helm charts automatically"
---

# Helm Charts
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

BeKind can automatically install Helm charts during cluster creation, making it easy to set up complex applications and services without manual intervention.

---

## Helm Stacks

{: .highlight }
**New in v0.6.0**: Organize and reuse Helm charts with stack configurations.

Helm Stacks allow you to organize related Helm charts into reusable configuration files. This is ideal for managing complex deployments with multiple interdependent charts.

### Creating a Stack

Stacks are stored in `~/.bekind/helmstack/<stack-name>/stack.yaml`:

```
~/.bekind/
└── helmstack/
    ├── argocd-cilium/
    │   └── stack.yaml
    ├── monitoring/
    │   └── stack.yaml
    └── database/
        └── stack.yaml
```

### Stack File Format

Each `stack.yaml` file contains a `helmCharts` array with the same format as inline chart definitions:

```yaml
helmCharts:
  - url: "https://helm.cilium.io"
    repo: "cilium"
    chart: "cilium"
    release: "cilium"
    namespace: "kube-system"
    wait: true
    valuesObject:
      kubeProxyReplacement: true
      operator:
        replicas: 1
      ingressController:
        enabled: true
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: true
    valuesObject:
      global:
        domain: argocd.example.com
```

### Using Stacks

Reference stacks in your BeKind configuration using the `helmStack` key:

```yaml
domain: "7f000001.nip.io"
helmStack:
  - name: argocd-cilium
  - name: monitoring
helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-rollouts"
    release: "argo-rollouts"
    namespace: "argo-rollouts"
    wait: true
kindConfig: |
  kind: Cluster
  name: my-cluster
  apiVersion: kind.x-k8s.io/v1alpha4
```

### Installation Order

Charts are installed in this order:

1. **Stack charts** (in the order stacks are listed)
2. **Inline charts** (from the `helmCharts` section)

Within each stack, charts are installed in the order they appear in the `stack.yaml` file.

### Benefits of Stacks

- **Reusability**: Share common chart configurations across multiple clusters
- **Organization**: Keep related charts together in logical groups
- **Maintainability**: Update stack definitions once, apply to many clusters
- **Modularity**: Mix and match stacks with inline charts as needed

---

## Configuration

Add Helm charts to your BeKind configuration under the `helmCharts` key:

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

## Configuration Options

### url (*Required*)

**Type**: `string`  
**Description**: The URL of the Helm repository. Can be a standard HTTP(S) URL or an OCI repository with `oci://` prefix.

```yaml
url: "https://kubernetes.github.io/ingress-nginx"
# or
url: "oci://registry.example.com/charts"
```

### repo (*Required*)

**Type**: `string`  
**Description**: The name to give the repository locally. This is equivalent to the `<reponame>` in `helm repo add <reponame> <url>`.

{: .note }
Ignored when using OCI repositories.

```yaml
repo: "ingress-nginx"
```

### chart (*Required*)

**Type**: `string`  
**Description**: The name of the chart to install from the Helm repository.

{: .note }
Ignored when using OCI repositories (full path included in URL).

```yaml
chart: "ingress-nginx"
```

### release (*Required*)

**Type**: `string`  
**Description**: The name to give the Helm release when it's installed.

```yaml
release: "nginx-ingress"
```

### namespace (*Required*)

**Type**: `string`  
**Description**: The Kubernetes namespace to install the release into. BeKind will create the namespace if it doesn't already exist.

```yaml
namespace: "ingress-controller"
```

### version

**Type**: `string`  
**Optional**: Yes  
**Description**: The specific version of the Helm chart to install. If not specified, the latest version is used.

```yaml
version: "4.8.3"
```

### wait

**Type**: `boolean`  
**Optional**: Yes  
**Default**: `false`  
**Description**: Whether to wait for the release to be fully installed before continuing. This is useful when subsequent steps depend on the chart being ready.

```yaml
wait: true
```

### valuesObject

**Type**: `object` (YAML)  
**Optional**: Yes  
**Description**: Custom values to pass to the Helm chart, equivalent to a values file. This allows you to customize the chart installation.

```yaml
valuesObject:
  controller:
    service:
      type: LoadBalancer
    replicas: 2
  resources:
    limits:
      cpu: "500m"
      memory: "512Mi"
```

---

## Examples

### Installing NGINX Ingress Controller

```yaml
helmCharts:
  - url: "https://kubernetes.github.io/ingress-nginx"
    repo: "ingress-nginx"
    chart: "ingress-nginx"
    release: "nginx-ingress"
    namespace: "ingress-nginx"
    wait: true
    valuesObject:
      controller:
        hostNetwork: true
        service:
          type: ClusterIP
```

### Installing Argo CD

```yaml
helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: true
    valuesObject:
      configs:
        secret:
          argocdServerAdminPassword: "$2a$10$..."
      server:
        ingress:
          enabled: true
          ingressClassName: nginx
```

### Installing Multiple Charts

```yaml
helmCharts:
  - url: "https://charts.jetstack.io"
    repo: "jetstack"
    chart: "cert-manager"
    release: "cert-manager"
    namespace: "cert-manager"
    version: "v1.13.0"
    wait: true
    valuesObject:
      installCRDs: true
      
  - url: "https://prometheus-community.github.io/helm-charts"
    repo: "prometheus-community"
    chart: "kube-prometheus-stack"
    release: "monitoring"
    namespace: "monitoring"
    wait: true
```

### Using OCI Registry

```yaml
helmCharts:
  - url: "oci://registry-1.docker.io/bitnamicharts/nginx"
    release: "my-nginx"
    namespace: "web"
```

---

## Important Notes

{: .warning }
**"Garbage in/garbage out"** - BeKind passes your Helm chart configuration directly to Helm. Invalid configurations will result in Helm errors. Always validate your chart values before using them with BeKind.

### Execution Order

Helm charts are installed in the order they appear in your configuration file. If you have dependencies between charts, list them in the correct order and use `wait: true` to ensure each chart is ready before the next one installs.

### Namespace Creation

BeKind automatically creates namespaces that don't exist. You don't need to create namespaces separately before installing charts.

### Values Validation

Test your `valuesObject` configuration with Helm before using it in BeKind:

```bash
# Test with a dry-run
helm install my-release repo/chart --dry-run --debug -f values.yaml
```

---

## Troubleshooting

### Chart Installation Fails

If a chart fails to install:

1. Check the Helm chart version is correct
2. Verify the repository URL is accessible
3. Validate your `valuesObject` against the chart's values schema
4. Review the BeKind output for Helm error messages

### Timeout Errors

If you get timeout errors:

1. Set `wait: true` and increase timeout if needed
2. Check if the chart requires specific node labels or taints
3. Verify sufficient resources are available in your cluster

### OCI Registry Issues

For OCI registries:

1. Ensure you're authenticated if the registry requires auth
2. Include the full chart path in the URL
3. Omit the `repo` and `chart` fields (not used with OCI)
