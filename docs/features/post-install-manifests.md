---
layout: default
title: Post Install Manifests
parent: Features
nav_order: 3
description: "Apply Kubernetes manifests automatically"
---

# Post Install Manifests
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

BeKind can automatically apply Kubernetes YAML manifests after the cluster is created and Helm charts are installed. This is useful for:

- Deploying applications
- Creating custom resources
- Setting up RBAC policies
- Configuring cluster-specific settings

---

## Configuration

Add manifest paths under the `postInstallManifests` key:

```yaml
postInstallManifests:
  - "file:///home/user/k8s/app.yaml"
  - "file:///home/user/k8s/service.yaml"
  - "file:///home/user/k8s/ingress.yaml"
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
postInstallManifests:
  - "file:///home/user/manifests/deployment.yaml"
  - "file:///home/user/manifests/service.yaml"
```

{: .warning }
Relative paths are not supported. Always use absolute paths with the `file://` prefix.

**Linux/macOS**:
```yaml
postInstallManifests:
  - "file:///home/username/k8s/app.yaml"
  - "file:///Users/username/projects/k8s/service.yaml"
```

**Windows**:
```yaml
postInstallManifests:
  - "file:///C:/Users/username/k8s/app.yaml"
```

### HTTP(S) URLs

Manifests can be fetched from remote URLs:

```yaml
postInstallManifests:
  - "https://yoursite.example.org/manifests/deployment.yaml"
  - "http://internal-server.local/configs/service.yaml"
```

{: .note }
Remote manifests must be served with the `text/plain` or `application/yaml` content type.

### Mixed Configuration

You can combine local files and remote URLs:

```yaml
postInstallManifests:
  - "file:///home/user/local/namespace.yaml"
  - "https://example.com/shared/rbac.yaml"
  - "file:///home/user/local/deployment.yaml"
  - "https://example.com/configs/service.yaml"
```

---

## Examples

### Single Application

```yaml
postInstallManifests:
  - "file:///home/user/apps/my-app/deployment.yaml"
  - "file:///home/user/apps/my-app/service.yaml"
  - "file:///home/user/apps/my-app/configmap.yaml"
```

### Multiple Applications

```yaml
postInstallManifests:
  - "file:///home/user/apps/frontend/all.yaml"
  - "file:///home/user/apps/backend/all.yaml"
  - "file:///home/user/apps/database/all.yaml"
  - "file:///home/user/networking/ingress.yaml"
```

### Remote Manifests

```yaml
postInstallManifests:
  - "https://raw.githubusercontent.com/user/repo/main/manifests/app.yaml"
  - "https://gist.githubusercontent.com/user/abc123/raw/deployment.yaml"
  - "https://yoursite.example.org/configs/service.yaml"
```

### Mixed Local and Remote

```yaml
postInstallManifests:
  - "file:///home/user/k8s/namespace.yaml"
  - "https://example.com/configs/base-config.yaml"
  - "file:///home/user/k8s/secrets.yaml"
  - "https://example.com/configs/monitoring.yaml"
```

### With Argo CD Applications

```yaml
helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: true

postInstallManifests:
  - "file:///home/user/argocd/app-of-apps.yaml"
  - "file:///home/user/argocd/projects.yaml"
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
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  key: value
```

**Multiple resources**:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: my-app
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: my-app
data:
  key: value
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: my-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: app
        image: my-app:latest
```

### Local File Accessibility

Ensure BeKind can read the manifest files:

```bash
# Check file exists
ls -l /home/user/k8s/app.yaml

# Check file permissions
chmod 644 /home/user/k8s/app.yaml
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
curl -I https://example.com/manifests/app.yaml

# Should return 200 OK with text/plain or application/yaml content type
```

---

## Execution Order

Manifests are applied in the order they appear in your configuration:

```yaml
postInstallManifests:
  - "file:///path/to/namespace.yaml"      # Applied first
  - "file:///path/to/configmap.yaml"      # Applied second
  - "file:///path/to/deployment.yaml"     # Applied third
```

{: .note }
If resources have dependencies, list them in the correct order. For example, create namespaces before resources that use them.

### Full Execution Flow

1. KIND cluster is created
2. Docker images are loaded (if configured)
3. Helm charts are installed (if configured)
4. **Post-install manifests are applied** ← You are here
5. Post-install actions are performed (if configured)

---

## Important Notes

### Error Handling

{: .warning }
**"Garbage in/garbage out"** - BeKind applies manifests directly using `kubectl apply`. Any errors come from the Kubernetes API server. There is no validation before application.

If a manifest fails to apply:
- BeKind will display the error
- Subsequent manifests will still be attempted
- The cluster will be in a partial state

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
   postInstallManifests:
     - "./manifests/app.yaml"
   
   # Correct
   postInstallManifests:
     - "file:///home/user/project/manifests/app.yaml"
   ```

2. **Verify the file exists**:
   ```bash
   ls -l /home/user/project/manifests/app.yaml
   ```

3. **Check file permissions**:
   ```bash
   chmod 644 /home/user/project/manifests/app.yaml
   ```

**For remote URLs:**

1. **Test URL accessibility**:
   ```bash
   curl -v https://example.com/manifests/app.yaml
   ```

2. **Check content type**:
   ```bash
   curl -I https://example.com/manifests/app.yaml
   # Should see: Content-Type: text/plain or application/yaml
   ```

3. **Verify network connectivity**:
   - Can you reach the server from your machine?
   - Are there firewall rules blocking access?
   - Is the URL correct (check for typos)?

### Application Failures

If `kubectl apply` fails:

1. **Validate YAML syntax**:
   ```bash
   yamllint manifest.yaml
   ```

2. **Check Kubernetes API version**:
   ```bash
   kubectl api-versions
   ```

3. **Verify resource definitions**:
   ```bash
   kubectl explain Deployment.spec
   ```

4. **Test manually**:
   ```bash
   kubectl apply -f manifest.yaml
   ```

### Namespace Issues

If resources fail because namespaces don't exist:

1. **Create namespace first**:
   ```yaml
   postInstallManifests:
     - "file:///path/to/namespace.yaml"
     - "file:///path/to/app.yaml"
   ```

2. **Or include namespace in manifest**:
   ```yaml
   apiVersion: v1
   kind: Namespace
   metadata:
     name: my-app
   ---
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: config
     namespace: my-app
   ```

### Dependency Ordering

If resources depend on each other:

1. **List dependencies first**:
   ```yaml
   postInstallManifests:
     - "file:///path/to/crd.yaml"           # Custom Resource Definition first
     - "file:///path/to/resource.yaml"      # Custom Resource second
   ```

2. **Use `wait` with Helm charts**: If manifests depend on Helm-installed resources:
   ```yaml
   helmCharts:
     - url: "..."
       wait: true  # Wait for chart to be ready
   
   postInstallManifests:
     - "file:///path/to/dependent-resource.yaml"
   ```

---

## Best Practices

### Use Remote Manifests for Shared Configurations

Store common manifests in a central location:

```yaml
postInstallManifests:
  # Shared base configuration
  - "https://config.company.com/k8s/base/namespace.yaml"
  - "https://config.company.com/k8s/base/rbac.yaml"
  # Local overrides
  - "file:///home/user/local/secrets.yaml"
  - "file:///home/user/local/app.yaml"
```

### Organize by Environment

```
.bekind/
├── profiles/
│   ├── dev/
│   │   ├── config.yaml
│   │   └── manifests/
│   │       ├── namespace.yaml
│   │       └── apps.yaml
│   └── staging/
│       ├── config.yaml
│       └── manifests/
│           └── apps.yaml
```

Reference in config:
```yaml
postInstallManifests:
  - "file://${HOME}/.bekind/profiles/dev/manifests/namespace.yaml"
  - "file://${HOME}/.bekind/profiles/dev/manifests/apps.yaml"
```

Or use remote URLs for consistency:
```yaml
postInstallManifests:
  - "https://github.com/company/k8s-configs/raw/main/dev/namespace.yaml"
  - "https://github.com/company/k8s-configs/raw/main/dev/apps.yaml"
```

### Combine Related Resources

Group related resources in single files:

```yaml
# app.yaml - everything for one application
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
---
apiVersion: v1
kind: Service
metadata:
  name: app
```

### Use with Post-Install Actions

Combine manifests with actions for complete setup:

```yaml
postInstallManifests:
  - "file:///home/user/k8s/app.yaml"

postInstallActions:
  - action: restart
    kind: Deployment
    name: app
    namespace: default
```
