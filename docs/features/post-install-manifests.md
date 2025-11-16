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

### File URLs

Each manifest must be specified as a `file://` URL with an absolute path:

```yaml
postInstallManifests:
  - "file:///home/user/manifests/deployment.yaml"
  - "file:///home/user/manifests/service.yaml"
```

{: .warning }
Relative paths are not supported. Always use absolute paths with the `file://` prefix.

### Path Examples

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

### File Accessibility

Ensure BeKind can read the manifest files:

```bash
# Check file exists
ls -l /home/user/k8s/app.yaml

# Check file permissions
chmod 644 /home/user/k8s/app.yaml
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

If BeKind can't find a manifest file:

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
