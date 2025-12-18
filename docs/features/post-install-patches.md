---
layout: default
title: Post Install Patches
parent: Features
nav_order: 5
description: "Apply JSON patches to Kubernetes resources"
---

# Post Install Patches
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

Post-install patches allow you to apply JSON Patch operations to Kubernetes resources after the cluster is created. This provides a Kustomize-like patching capability for fine-grained modifications without needing to manage entire manifest files.

This is useful for:

- Modifying Gateway API routes after Helm installation
- Adjusting resource configurations dynamically
- Fixing port numbers or other specific values
- Updating annotations or labels on existing resources

Patches run **after** Helm charts are installed and manifests are applied, but **before** post-install actions are executed.

---

## Configuration

Add patches under the `postInstallPatches` key:

```yaml
postInstallPatches:
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: GRPCRoute
      name: argocd-server-grpc
      namespace: argocd
    patch: |
      - op: replace
        path: /spec/rules/0/backendRefs/0/port
        value: 443
```

---

## Configuration Options

### target (*Required*)

Specifies the Kubernetes resource to patch.

#### target.version (*Required*)

**Type**: `string`  
**Description**: The API version of the resource.

```yaml
target:
  version: v1
```

#### target.kind (*Required*)

**Type**: `string`  
**Description**: The kind of the resource (e.g., `Service`, `Deployment`, `HTTPRoute`).

```yaml
target:
  kind: GRPCRoute
```

#### target.name (*Required*)

**Type**: `string`  
**Description**: The name of the resource to patch.

```yaml
target:
  name: argocd-server-grpc
```

#### target.group (*Optional*)

**Type**: `string`  
**Default**: `""` (core API group)  
**Description**: The API group of the resource. If empty or not provided, defaults to the core API group.

```yaml
# For core resources like Pod, Service, ConfigMap
target:
  group: ""
  # or omit entirely
  
# For custom resources
target:
  group: gateway.networking.k8s.io
```

#### target.namespace (*Optional*)

**Type**: `string`  
**Default**: `"default"`  
**Description**: The namespace where the resource exists. If not provided, defaults to `default`.

```yaml
target:
  namespace: argocd
```

### patch (*Required*)

**Type**: `string` (YAML array)  
**Description**: A JSON Patch (RFC 6902) document in YAML format specifying the operations to apply.

```yaml
patch: |
  - op: replace
    path: /spec/rules/0/backendRefs/0/port
    value: 443
```

---

## JSON Patch Operations

The `patch` field accepts a JSON array of operations following [RFC 6902](https://tools.ietf.org/html/rfc6902).

### Supported Operations

#### add

Adds a value at the specified path.

```yaml
patch: |
  - op: add
    path: /metadata/labels/app
    value: my-app
```

#### replace

Replaces an existing value at the specified path.

```yaml
patch: |
  - op: replace
    path: /spec/replicas
    value: 3
```

#### remove

Removes the value at the specified path.

```yaml
patch: |
  - op: remove
    path: /metadata/annotations/old-annotation
```

#### copy

Copies a value from one path to another.

```yaml
patch: |
  - op: copy
    from: /spec/template/metadata/labels
    path: /metadata/labels
```

#### move

Moves a value from one path to another.

```yaml
patch: |
  - op: move
    from: /spec/oldField
    path: /spec/newField
```

#### test

Tests that a value at a path equals a specified value (fails the patch if not equal).

```yaml
patch: |
  - op: test
    path: /spec/replicas
    value: 1
  - op: replace
    path: /spec/replicas
    value: 3
```

---

## Examples

### Example 1: Fixing Gateway API Backend Port

Modify a GRPCRoute to change the backend port from 80 to 443:

```yaml
postInstallPatches:
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: GRPCRoute
      name: argocd-server-grpc
      namespace: argocd
    patch: |
      - op: replace
        path: /spec/rules/0/backendRefs/0/port
        value: 443
```

### Example 2: Adding Labels to a Service

Add application labels to an existing service in the default namespace:

```yaml
postInstallPatches:
  - target:
      version: v1
      kind: Service
      name: my-service
      # namespace defaults to "default"
      # group defaults to "" (core)
    patch: |
      - op: add
        path: /metadata/labels/app
        value: my-app
      - op: add
        path: /metadata/labels/version
        value: v1.0.0
```

### Example 3: Updating Multiple HTTPRoute Backends

Change both HTTP and HTTPS backend ports for an HTTPRoute:

```yaml
postInstallPatches:
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: HTTPRoute
      name: argocd-server
      namespace: argocd
    patch: |
      - op: replace
        path: /spec/rules/0/backendRefs/0/port
        value: 80
```

### Example 4: Modifying Deployment Replicas

Adjust the replica count of a deployment:

```yaml
postInstallPatches:
  - target:
      group: apps
      version: v1
      kind: Deployment
      name: nginx-deployment
      namespace: production
    patch: |
      - op: replace
        path: /spec/replicas
        value: 5
```

### Example 5: Adding Annotations to a Core Resource

Add annotations to a ConfigMap in the core API group:

```yaml
postInstallPatches:
  - target:
      group: ""  # Explicitly specify core group
      version: v1
      kind: ConfigMap
      name: app-config
      namespace: default
    patch: |
      - op: add
        path: /metadata/annotations
        value:
          description: Application configuration
          managed-by: bekind
```

### Example 6: Multiple Patches

Apply multiple patches to different resources:

```yaml
postInstallPatches:
  # Fix GRPCRoute port
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: GRPCRoute
      name: argocd-server-grpc
      namespace: argocd
    patch: |
      - op: replace
        path: /spec/rules/0/backendRefs/0/port
        value: 443
  
  # Fix HTTPRoute port
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: HTTPRoute
      name: argocd-server
      namespace: argocd
    patch: |
      - op: replace
        path: /spec/rules/0/backendRefs/0/port
        value: 80
  
  # Add label to service
  - target:
      version: v1
      kind: Service
      name: argocd-server
      namespace: argocd
    patch: |
      - op: add
        path: /metadata/labels/patched
        value: "true"
```

---

## Complete Configuration Example

Here's a complete BeKind configuration using post-install patches with Argo CD and Gateway API:

```yaml
kindConfig: |
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  name: argocd-gateway
  nodes:
    - role: control-plane

helmCharts:
  - url: https://cilium.github.io/charts
    repo: cilium
    chart: cilium
    release: cilium
    namespace: kube-system
    wait: true
    valuesObject:
      kubeProxyReplacement: true
      gatewayAPI:
        enabled: true

  - url: https://argoproj.github.io/argo-helm
    repo: argo
    chart: argo-cd
    release: argocd
    namespace: argocd
    wait: true
    valuesObject:
      server:
        service:
          type: ClusterIP

postInstallManifests:
  - "file:///path/to/gateway.yaml"
  - "file:///path/to/httproute.yaml"
  - "file:///path/to/grpcroute.yaml"

postInstallPatches:
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: GRPCRoute
      name: argocd-server-grpc
      namespace: argocd
    patch: |
      - op: replace
        path: /spec/rules/0/backendRefs/0/port
        value: 443
  
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: HTTPRoute
      name: argocd-server
      namespace: argocd
    patch: |
      - op: replace
        path: /spec/rules/0/backendRefs/0/port
        value: 80
```

---

## Execution Order

Post-install patches execute in the following order:

1. KIND cluster creation
2. Helm chart installations
3. Post-install manifests
4. **Post-install patches** ← You are here
5. Post-install actions
6. BeKind config saved to secret

This ensures patches can modify resources created by earlier steps, and then actions can be performed on the patched resources (e.g., restarting deployments after patching their configuration).

---

## Error Handling

If a patch fails:
- A warning is logged
- Execution continues with remaining patches
- The cluster setup completes

Common failure reasons:
- Resource not found (name or namespace incorrect)
- Invalid patch syntax
- Path doesn't exist in the resource
- API group/version/kind not recognized

---

## Tips and Best Practices

### 1. Use Valid YAML Syntax

Ensure your patch is valid YAML. The patch field accepts YAML format for easier readability:

```yaml
patch: |
  - op: replace
    path: /spec/replicas
    value: 3
```

### 2. Test Patches Manually First

Before adding to BeKind config, test patches manually using JSON format with kubectl:

```bash
kubectl patch grpcroute argocd-server-grpc -n argocd --type='json' \
  -p='[{"op": "replace", "path": "/spec/rules/0/backendRefs/0/port", "value": 443}]'
```

### 3. Use Descriptive Comments

Add comments in your YAML to explain why patches are needed:

```yaml
postInstallPatches:
  # Argo CD Helm chart creates routes with wrong port, fix the backend port
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: GRPCRoute
      name: argocd-server-grpc
      namespace: argocd
    patch: |
      - op: replace
        path: /spec/rules/0/backendRefs/0/port
        value: 443
```

### 4. Array Indexing

JSON Patch uses zero-based array indexing. The first element is `0`:

```yaml
# Correct
path: "/spec/rules/0/backendRefs/0/port"

# Incorrect
path: "/spec/rules/1/backendRefs/1/port"  # This is the second element
```

### 5. Escape Special Characters

If paths contain special characters, use JSON Pointer escaping:
- `~` becomes `~0`
- `/` becomes `~1`

```yaml
# To reference key "app/version"
path: "/metadata/labels/app~1version"
```

### 6. Verify Resource Existence

Ensure resources exist before patching. Check logs during execution:

```
INFO Post Install Patches
INFO Applying patch to GRPCRoute/argocd-server-grpc in namespace argocd
INFO Successfully patched GRPCRoute/argocd-server-grpc
```

---

## Comparison with Other Tools

### vs. Kustomize

**Similarities:**
- JSON Patch support (RFC 6902)
- Target specific resources
- Non-destructive modifications

**Differences:**
- BeKind patches run at cluster creation time
- No need for separate Kustomize overlays
- Integrated into single configuration file

### vs. Post-Install Actions

| Feature | Post-Install Patches | Post-Install Actions |
|---------|---------------------|---------------------|
| Purpose | Modify resource fields | Restart/delete resources |
| Granularity | Field-level changes | Resource-level operations |
| Operations | add, replace, remove, etc. | restart, delete |
| Use case | Fix configurations | Trigger rollouts |

---

## Troubleshooting

### Patch Not Applied

**Symptom**: No changes visible after cluster creation.

**Solutions**:
1. Check logs for warnings during patch execution
2. Verify resource exists: `kubectl get <kind> <name> -n <namespace>`
3. Verify API group/version is correct: `kubectl api-resources`
4. Test patch manually with `kubectl patch`

### Invalid JSON Syntax

**Symptom**: Warning: "Issue with Post Install Patches"

**Solutions**:
1. Validate JSON with `jq` or online validator
2. Ensure proper escaping in YAML (use `|` for multiline)
3. Check for missing commas or brackets

### Path Not Found

**Symptom**: Patch fails with "path not found" error

**Solutions**:
1. Verify path exists: `kubectl get <kind> <name> -n <namespace> -o json | jq .spec.rules[0].backendRefs[0].port`
2. Check array indices (zero-based)
3. Ensure parent paths exist before adding nested values

### Permission Denied

**Symptom**: Patch fails with authorization error

**Solutions**:
1. Ensure kubeconfig has proper permissions
2. Check RBAC for patch operations on target resources
3. Verify namespace access

---

## Related Features

- [Post Install Manifests](post-install-manifests.md) - Apply complete manifests
- [Post Install Actions](post-install-actions.md) - Restart or delete resources
- [Helm Charts](helm-charts.md) - Install applications with Helm

---

## Additional Resources

- [RFC 6902: JSON Patch](https://tools.ietf.org/html/rfc6902)
- [Kubernetes Strategic Merge Patch](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/)
- [kubectl patch documentation](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#patch)
