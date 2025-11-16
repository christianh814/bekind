---
layout: default
title: Loading Docker Images
parent: Features
nav_order: 2
description: "Pre-load Docker images into KIND clusters"
---

# Loading Docker Images
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

BeKind can pre-load Docker images into your KIND cluster nodes. This is useful for:

- Avoiding image pull delays during testing
- Working with private or locally-built images
- Ensuring specific image versions are available
- Speeding up pod startup times

{: .note }
Only `docker` is supported. See the [KIND upstream issue](https://github.com/kubernetes-sigs/kind/issues/2038) for more information.

---

## Configuration

Add image loading configuration under the `loadDockerImages` key:

```yaml
loadDockerImages:
  pullImages: true
  images:
    - gcr.io/kuar-demo/kuard-amd64:blue
    - quay.io/christianh814/simple-go:latest
    - my-local-image:dev
```

---

## Configuration Options

### pullImages

**Type**: `boolean`  
**Optional**: Yes  
**Default**: `true`  
**Description**: Whether to pull images from the registry before loading them into the cluster. This is a global setting that applies to all images in the list.

```yaml
pullImages: true
```

Set to `false` if:
- All images already exist locally
- You're using locally-built images that aren't in a registry
- You want to speed up cluster creation (but risk missing images)

### images

**Type**: `array` of `string`  
**Optional**: Yes  
**Description**: List of Docker images to load into the cluster. Each image should be specified with its full name and tag.

```yaml
images:
  - nginx:latest
  - postgres:15-alpine
  - gcr.io/my-project/my-app:v1.2.3
```

---

## Examples

### Load Public Images

```yaml
loadDockerImages:
  pullImages: true
  images:
    - nginx:1.25
    - redis:7-alpine
    - postgres:15
```

### Load Images from Different Registries

```yaml
loadDockerImages:
  pullImages: true
  images:
    - docker.io/library/nginx:latest
    - gcr.io/kuar-demo/kuard-amd64:blue
    - quay.io/prometheus/prometheus:v2.45.0
    - ghcr.io/my-org/my-app:main
```

### Load Local Images Without Pulling

```yaml
loadDockerImages:
  pullImages: false
  images:
    - my-app:dev
    - my-api:test
```

### Mixed Scenario

If you need to pull some images but not others, you'll need to:

1. Build/pull all images locally first
2. Set `pullImages: false`
3. List all images

```bash
# Pull images you need
docker pull nginx:latest
docker pull redis:alpine

# Build local images
docker build -t my-app:dev .

# Then use BeKind with pullImages: false
```

---

## How It Works

1. **Pull Phase** (if `pullImages: true`):
   - BeKind uses `docker pull` to fetch each image from its registry
   - Images are downloaded to your local Docker image cache

2. **Load Phase**:
   - BeKind uses KIND's `kind load docker-image` command
   - Images are copied from your local Docker to the KIND cluster nodes
   - Images become available to pods without needing to pull from registries

---

## Important Notes

### Docker Only

{: .warning }
Currently, only Docker is supported for loading images. Podman and other container runtimes are not supported due to KIND limitations.

### Image Tags

Always specify image tags explicitly. Avoid using `:latest` in production configurations, as it can lead to inconsistent environments.

```yaml
# Good
images:
  - nginx:1.25.3

# Avoid in production
images:
  - nginx:latest
```

### Image Size

Loading large images can take time and consume disk space. Consider:
- Using smaller base images (alpine variants)
- Multi-stage builds to reduce final image size
- Loading only necessary images

### Execution Timing

Images are loaded **after** the cluster is created but **before** Helm charts are installed and manifests are applied. This ensures images are available when pods are created.

---

## Troubleshooting

### Pull Failures

If image pulls fail:

1. **Check image name and tag are correct**:
   ```bash
   docker pull <image-name>
   ```

2. **Verify registry access**:
   - Public registries should work without authentication
   - Private registries require Docker login:
     ```bash
     docker login <registry-url>
     ```

3. **Check network connectivity**:
   - Ensure you can reach the registry
   - Check firewall rules if behind a corporate network

### Load Failures

If loading images into KIND fails:

1. **Verify Docker is running**:
   ```bash
   docker ps
   ```

2. **Check KIND cluster exists**:
   ```bash
   kind get clusters
   ```

3. **Manually test loading**:
   ```bash
   kind load docker-image <image-name> --name <cluster-name>
   ```

### Images Not Available in Pods

If pods can't find images after loading:

1. **Check imagePullPolicy**: Set to `IfNotPresent` or `Never` in pod specs:
   ```yaml
   spec:
     containers:
     - name: myapp
       image: my-app:dev
       imagePullPolicy: IfNotPresent
   ```

2. **Verify image was loaded**:
   ```bash
   docker exec <kind-node> crictl images
   ```

3. **Check image name matches exactly**: Tag and registry must match precisely

---

## Best Practices

### Development Workflow

For local development:

```yaml
loadDockerImages:
  pullImages: false  # You're building locally
  images:
    - my-app:dev
    - my-api:dev
```

Build your images before starting the cluster:

```bash
docker build -t my-app:dev .
bekind start --profile dev
```

### CI/CD Workflow

For CI/CD pipelines:

```yaml
loadDockerImages:
  pullImages: true  # Always pull to ensure latest
  images:
    - my-app:${CI_COMMIT_SHA}
    - postgres:15-alpine
    - redis:7-alpine
```

### Testing Specific Versions

When testing version compatibility:

```yaml
loadDockerImages:
  pullImages: true
  images:
    - postgres:14-alpine
    - postgres:15-alpine
    - postgres:16-alpine
```
