---
layout: default
title: Pre Helm Manifests
parent: Features
nav_order: 3
description: "Apply Kubernetes manifests before Helm charts"
---

# Pre Helm Manifests
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

BeKind can automatically apply Kubernetes YAML manifests **before** Helm charts are installed. This is useful for creating resources that your charts depend on, such as:

- Namespaces
- Custom Resource Definitions (CRDs)
- Secrets and ConfigMaps consumed by charts
- RBAC policies required during install
- Any prerequisite resources a chart expects to already exist

This works just like [Post Install Manifests]({% link features/post-install-manifests.md %}), but the manifests are applied earlier in the setup flow, before any Helm chart is installed.

---

## Configuration

Add manifest paths under the `preHelmManifests` key:

```yaml
preHelmManifests:
  - "file:///home/user/k8s/namespace.yaml"
  - "file:///home/user/k8s/crds.yaml"
  - "https://example.com/configs/secret.yaml"
```

---

## Configuration Format

### Supported URL Formats

BeKind supports multiple ways to specify manifest locations:

- **Local files**: `file://` URLs with absolute paths
- **HTTP(S) URLs**: Direct links to manifests served as `text/plain`

You can mix and match both types in the same configuration.

### File URLs

Local files must be specified as `file://` URLs with absolute paths:

```yaml
preHelmManifests:
  - "file:///home/user/manifests/namespace.yaml"
  - "file:///home/user/manifests/crds.yaml"
```

{: .warning }
Relative paths are not supported. Always use absolute paths with the `file://` prefix.

**Linux/macOS**:
```yaml
preHelmManifests:
  - "file:///home/username/k8s/namespace.yaml"
  - "file:///Users/username/projects/k8s/crds.yaml"
```

**Windows**:
```yaml
preHelmManifests:
  - "file:///C:/Users/username/k8s/namespace.yaml"
```

### HTTP(S) URLs

Manifests can be fetched from remote URLs:

```yaml
preHelmManifests:
  - "https://yoursite.example.org/manifests/crds.yaml"
  - "http://internal-server.local/configs/namespace.yaml"
```

{: .note }
Remote manifests must be served with the `text/plain` or `application/yaml` content type.

### Mixed Configuration

You can combine local files and remote URLs:

```yaml
preHelmManifests:
  - "file:///home/user/local/namespace.yaml"
  - "https://example.com/shared/crds.yaml"
  - "file:///home/user/local/secret.yaml"
```

---

## Examples

### Create a Namespace Before Installing a Chart

```yaml
preHelmManifests:
  - "file:///home/user/k8s/monitoring-namespace.yaml"

helmCharts:
  - url: "https://prometheus-community.github.io/helm-charts"
    repo: "prometheus-community"
    chart: "kube-prometheus-stack"
    release: "monitoring"
    namespace: "monitoring"
    wait: true
```

### Install CRDs Before a Chart That Requires Them

```yaml
preHelmManifests:
  - "https://example.com/crds/my-operator-crds.yaml"

helmCharts:
  - url: "https://example.com/charts"
    repo: "myrepo"
    chart: "my-operator"
    release: "my-operator"
    namespace: "operators"
    wait: true
```

### Provide a Secret Consumed by a Chart

```yaml
preHelmManifests:
  - "file:///home/user/k8s/registry-credentials.yaml"

helmCharts:
  - url: "https://charts.example.com"
    repo: "example"
    chart: "private-app"
    release: "private-app"
    namespace: "default"
    wait: true
```

---

## Manifest Requirements

### YAML Format

Manifests must be valid Kubernetes YAML files. They can contain:

- Single resources
- Multiple resources separated by `---`
- Any valid Kubernetes resource type

**Single resource**:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
```

**Multiple resources**:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: my-app
---
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: my-app
type: Opaque
stringData:
  token: "example"
```

### Namespaces

For namespace-scoped resources, BeKind honors the `metadata.namespace` field in
the manifest. If a namespace-scoped resource omits `metadata.namespace`, BeKind
applies it to the `default` namespace (mirroring `kubectl` behavior).

{: .note }
Cluster-scoped resources (such as `Namespace`, `ClusterRole`, or CRDs) are
unaffected by this defaulting.

### Local File Accessibility

Ensure BeKind can read the manifest files:

```bash
# Check file exists
ls -l /home/user/k8s/namespace.yaml

# Check file permissions
chmod 644 /home/user/k8s/namespace.yaml
```

### Remote URL Accessibility

For HTTP(S) URLs, ensure:
- The URL is accessible from where BeKind is running
- The server returns `text/plain` or `application/yaml` content type
- No authentication is required, or use a URL with embedded credentials (not recommended for production)
- HTTPS certificates are valid (or use HTTP for internal/trusted networks)

Test URL accessibility:

```bash
# Test with curl
curl -I https://example.com/manifests/crds.yaml

# Should return 200 OK with text/plain or application/yaml content type
```

---

## Execution Order

Manifests are applied in the order they appear in your configuration:

```yaml
preHelmManifests:
  - "file:///path/to/namespace.yaml"      # Applied first
  - "file:///path/to/crds.yaml"           # Applied second
  - "file:///path/to/secret.yaml"         # Applied third
```

{: .note }
If resources have dependencies, list them in the correct order. For example, create namespaces before resources that use them.

### Full Execution Flow

1. KIND cluster is created
2. Docker images are loaded (if configured)
3. **Pre-helm manifests are applied** ← You are here
4. Helm charts are installed (if configured)
5. Post-install manifests are applied (if configured)
6. Post-install patches are applied (if configured)
7. Post-install actions are performed (if configured)

This ordering lets you stage prerequisite resources (namespaces, CRDs, secrets) so that Helm charts installed afterward can rely on them.

---

## Important Notes

### Error Handling

Pre-helm manifests are applied on a **best-effort** basis. If a manifest fails to apply:

- BeKind logs a warning
- Subsequent manifests are still attempted
- Helm chart installation continues
- The cluster setup is not aborted

{: .warning }
Because application is best-effort, BeKind will continue even if a prerequisite manifest fails. If a chart strictly depends on a resource created here, verify it was applied successfully.

### Validation Before Use

Always validate your manifests before using them with BeKind:

```bash
# Dry-run validation
kubectl apply -f manifest.yaml --dry-run=client

# Server-side validation
kubectl apply -f manifest.yaml --dry-run=server
```

### Supported Formats

Currently, only YAML files are supported. JSON manifests are not supported.

---

## Troubleshooting

### File Not Found

If BeKind can't find a manifest:

**For local files:**

1. **Check the path is absolute**:
   ```yaml
   # Wrong
   preHelmManifests:
     - "./manifests/namespace.yaml"

   # Correct
   preHelmManifests:
     - "file:///home/user/project/manifests/namespace.yaml"
   ```

2. **Verify the file exists**:
   ```bash
   ls -l /home/user/project/manifests/namespace.yaml
   ```

**For remote URLs:**

1. **Test URL accessibility**:
   ```bash
   curl -v https://example.com/manifests/crds.yaml
   ```

2. **Check content type**:
   ```bash
   curl -I https://example.com/manifests/crds.yaml
   # Should see: Content-Type: text/plain or application/yaml
   ```

### Application Failures

If a manifest fails to apply:

1. **Validate YAML syntax**:
   ```bash
   yamllint manifest.yaml
   ```

2. **Check Kubernetes API version**:
   ```bash
   kubectl api-versions
   ```

3. **Test manually**:
   ```bash
   kubectl apply -f manifest.yaml
   ```

### Chart Still Fails After Adding a Prerequisite

Because pre-helm manifests are best-effort, a failed prerequisite does not stop the chart install:

1. Re-run BeKind with debug logging to see the warning:
   ```bash
   bekind start --config config.yaml -v
   ```

2. Confirm the resource was created:
   ```bash
   kubectl get ns
   kubectl get crds
   ```

---

## Best Practices

### Stage Namespaces and CRDs Here

Use `preHelmManifests` for resources charts expect to already exist:

```yaml
preHelmManifests:
  - "file:///home/user/k8s/namespaces.yaml"
  - "https://example.com/crds/operator-crds.yaml"

helmCharts:
  - url: "https://example.com/charts"
    repo: "example"
    chart: "operator"
    release: "operator"
    namespace: "operators"
    wait: true
```

### Pair with Post Install Manifests

Use `preHelmManifests` for prerequisites and [`postInstallManifests`]({% link features/post-install-manifests.md %}) for resources that depend on the installed charts:

```yaml
preHelmManifests:
  - "file:///home/user/k8s/namespace.yaml"

helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: true

postInstallManifests:
  - "file:///home/user/argocd/app-of-apps.yaml"
```
