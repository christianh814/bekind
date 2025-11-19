---
layout: default
title: Post Install Actions
parent: Features
nav_order: 4
description: "Automate resource restarts and deletions"
---

# Post Install Actions
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

Post-install actions allow you to perform automated operations on Kubernetes resources after the cluster setup is complete. This is useful for:

- Restarting deployments after configuration changes
- Cleaning up temporary pods
- Forcing rollouts to pick up new configurations
- Removing debug resources

Actions run **after** Helm charts are installed, manifests are applied, and patches are applied.

---

## Supported Actions

### restart

Performs a rollout restart on the specified resource (equivalent to `kubectl rollout restart`).

**Supported resource kinds**:
- `Deployment`
- `StatefulSet`
- `DaemonSet`

### delete

Deletes the specified resource (equivalent to `kubectl delete`).

**Supported resource kinds**:
- `Pod`

---

## Configuration

Add actions under the `postInstallActions` key:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    name: argocd-server
    namespace: argocd
    
  - action: delete
    kind: Pod
    name: old-pod
    namespace: default
```

---

## Configuration Options

### action (*Required*)

**Type**: `string`  
**Values**: `restart` | `delete`  
**Description**: The action to perform on the resource.

```yaml
action: restart
```

### kind (*Required*)

**Type**: `string`  
**Description**: The kind of Kubernetes resource.

- For `restart` action: `Deployment`, `StatefulSet`, or `DaemonSet`
- For `delete` action: `Pod`

```yaml
kind: Deployment
```

### namespace

**Type**: `string`  
**Optional**: Yes  
**Default**: `default`  
**Description**: The namespace where the resource(s) exist.

```yaml
namespace: argocd
```

### name

**Type**: `string`  
**Optional**: Either `name` or `labelSelector` must be provided  
**Description**: The name of a specific resource to target.

```yaml
name: argocd-server
```

{: .note }
If both `name` and `labelSelector` are provided, `labelSelector` takes precedence and `name` is ignored.

### labelSelector

**Type**: `object` (map of label key-value pairs)  
**Optional**: Either `name` or `labelSelector` must be provided  
**Description**: Labels to select multiple resources. All resources matching the labels will be affected.

```yaml
labelSelector:
  app.kubernetes.io/name: argocd-server
  app.kubernetes.io/component: server
```

### group

**Type**: `string`  
**Optional**: Yes  
**Description**: The API group of the resource.

**Defaults**:
- `Pod`: `` (empty string / core API)
- `Deployment`, `StatefulSet`, `DaemonSet`: `apps`

You typically don't need to specify this unless working with custom resources.

```yaml
group: apps
```

### version

**Type**: `string`  
**Optional**: Yes  
**Default**: `v1`  
**Description**: The API version of the resource.

```yaml
version: v1
```

---

## Examples

### Restart by Name

Restart a specific deployment by name:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    name: argocd-server
    namespace: argocd
```

This is equivalent to:
```bash
kubectl rollout restart deployment/argocd-server -n argocd
```

### Restart by Label Selector

Restart all deployments matching specific labels:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    namespace: argocd
    labelSelector:
      app.kubernetes.io/name: argocd-applicationset-controller
```

This is equivalent to:
```bash
kubectl rollout restart deployment -n argocd -l app.kubernetes.io/name=argocd-applicationset-controller
```

### Restart Multiple Resources

Use multiple labels to target specific resources:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    namespace: production
    labelSelector:
      app: myapp
      environment: production
      version: v2
```

### Restart Different Resource Types

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    name: web-app
    namespace: default
    
  - action: restart
    kind: StatefulSet
    name: database
    namespace: default
    
  - action: restart
    kind: DaemonSet
    name: monitoring-agent
    namespace: monitoring
```

### Delete by Name

Delete a specific pod:

```yaml
postInstallActions:
  - action: delete
    kind: Pod
    name: temporary-job
    namespace: default
```

This is equivalent to:
```bash
kubectl delete pod/temporary-job -n default
```

### Delete by Label Selector

Delete all pods matching specific labels:

```yaml
postInstallActions:
  - action: delete
    kind: Pod
    namespace: kube-system
    labelSelector:
      app: cleanup
      temporary: "true"
```

This is equivalent to:
```bash
kubectl delete pod -n kube-system -l app=cleanup,temporary=true
```

### Combined Workflow

Restart deployments and clean up temporary pods:

```yaml
postInstallActions:
  # Restart Argo CD components to pick up new config
  - action: restart
    kind: Deployment
    namespace: argocd
    labelSelector:
      app.kubernetes.io/part-of: argocd
  
  # Remove debug pods
  - action: delete
    kind: Pod
    namespace: default
    labelSelector:
      debug: "true"
  
  # Remove specific temporary pod
  - action: delete
    kind: Pod
    name: init-job-xyz
    namespace: default
```

---

## How It Works

### Restart Action

When you restart a resource, BeKind:

1. Gets the resource using the Kubernetes dynamic client
2. Updates the pod template annotation:
   ```yaml
   kubectl.kubernetes.io/restartedAt: "2024-11-16T10:30:00Z"
   ```
3. This triggers Kubernetes to perform a rolling restart

The restart is graceful and follows the resource's update strategy.

### Delete Action

When you delete a resource, BeKind:

1. Finds the resource(s) matching the criteria
2. Deletes each resource using the Kubernetes API
3. The resource is removed immediately (subject to grace periods)

---

## Validation

BeKind validates your configuration before executing:

### Required Fields

- `action` must be specified
- `kind` must be specified
- Either `name` or `labelSelector` must be provided

### Action-Kind Compatibility

- `restart` action only works with: `Deployment`, `StatefulSet`, `DaemonSet`
- `delete` action only works with: `Pod`

Using invalid combinations will result in an error.

---

## Execution Order

Actions are executed in the order they appear in your configuration:

```yaml
postInstallActions:
  - action: restart
    kind: Deployment
    name: first         # Executed first
    
  - action: restart
    kind: Deployment
    name: second        # Executed second
    
  - action: delete
    kind: Pod
    name: cleanup       # Executed third
```

### Full Execution Flow

1. KIND cluster is created
2. Docker images are loaded (if configured)
3. Helm charts are installed (if configured)
4. Post-install manifests are applied (if configured)
5. Post-install patches are applied (if configured)
6. **Post-install actions are performed** ← You are here

This order allows you to patch resources first, then trigger actions like restarts to pick up the patched configurations.

---

## Important Notes

### Label Selector Precedence

{: .warning }
If both `name` and `labelSelector` are provided, `labelSelector` takes precedence and `name` is ignored.

```yaml
# This will use labelSelector and ignore name
postInstallActions:
  - action: restart
    kind: Deployment
    name: ignored-name                    # This is ignored
    labelSelector:
      app: myapp                          # This is used
```

### Core API Group

For Pods (core API resources), use an empty string for the group or omit it entirely:

```yaml
# Both are correct
postInstallActions:
  - action: delete
    kind: Pod
    group: ""           # Explicit empty string
    
  - action: delete
    kind: Pod           # Group omitted (uses default)
```

### No Rollback

Actions are executed once and not tracked. There is no automatic rollback if something fails.

---

## Troubleshooting

### Resource Not Found

If you get "resource not found" errors:

1. **Check resource exists**:
   ```bash
   kubectl get deployment argocd-server -n argocd
   ```

2. **Verify namespace**:
   ```bash
   kubectl get deployment --all-namespaces | grep argocd-server
   ```

3. **Check spelling**: Resource names and namespaces are case-sensitive

### Label Selector Not Matching

If label selectors don't find resources:

1. **Check labels on resources**:
   ```bash
   kubectl get deployment -n argocd --show-labels
   ```

2. **Verify label syntax**:
   ```yaml
   # Correct
   labelSelector:
     app.kubernetes.io/name: argocd
   
   # Wrong - values must be strings
   labelSelector:
     replicas: 1  # This won't work
   ```

3. **Test selector manually**:
   ```bash
   kubectl get deployment -n argocd -l app.kubernetes.io/name=argocd
   ```

### Action Not Supported

If you get "action not supported" errors:

1. **Check action name**: Must be exactly `restart` or `delete`

2. **Verify kind is supported**:
   - `restart`: Only `Deployment`, `StatefulSet`, `DaemonSet`
   - `delete`: Only `Pod`

### Permission Errors

If you get permission errors:

1. **Check cluster access**:
   ```bash
   kubectl auth can-i update deployments -n argocd
   kubectl auth can-i delete pods -n default
   ```

2. **Verify RBAC**: Ensure your kubeconfig has appropriate permissions

---

## Best Practices

### Use with Helm Charts

Restart deployments after Helm installs to pick up new configurations:

```yaml
helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: true

postInstallActions:
  - action: restart
    kind: Deployment
    namespace: argocd
    labelSelector:
      app.kubernetes.io/part-of: argocd
```

### Clean Up After Manifests

Remove temporary resources created by manifests:

```yaml
postInstallManifests:
  - "file:///path/to/init-job.yaml"

postInstallActions:
  - action: delete
    kind: Pod
    namespace: default
    labelSelector:
      job: init
```

### Use Label Selectors for Flexibility

Label selectors are more flexible than names:

```yaml
# Good - works even if deployment name changes
postInstallActions:
  - action: restart
    kind: Deployment
    labelSelector:
      app: myapp

# Less flexible - breaks if name changes
postInstallActions:
  - action: restart
    kind: Deployment
    name: myapp-deployment-v1
```

### Validate Labels First

Before using label selectors in BeKind, test them:

```bash
# Check what will be affected
kubectl get deployment -n argocd -l app.kubernetes.io/part-of=argocd

# Verify it's what you expect
kubectl get deployment -n argocd -l app.kubernetes.io/part-of=argocd -o name
```
