---
layout: default
title: Quick Start
nav_order: 2
description: "Get started with BeKind in minutes"
---

# Quick Start
{: .no_toc }

Get up and running with BeKind in just a few minutes.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Prerequisites

Before starting, make sure you have:
- [Docker](https://docs.docker.com/get-docker/) installed and running
- [kubectl](https://kubernetes.io/docs/tasks/tools/) installed
- BeKind installed (see [Installation]({% link installation.md %}))

---

## Create a Configuration File

Create a configuration file at `/tmp/config.yaml`:

```bash
cat << 'EOF' > /tmp/config.yaml
helmCharts:
  - url: "oci://ghcr.io/stefanprodan/charts/podinfo"
    repo: "podinfo"
    chart: "podinfo"
    release: "podinfo"
    namespace: "demo"
    wait: true
    valuesObject:
      replicaCount: 2
      image:
        pullPolicy: "IfNotPresent"
      ingress:
        enabled: "false"
kindConfig: |
  kind: Cluster
  name: mycluster
  apiVersion: kind.x-k8s.io/v1alpha4
  nodes:
  - role: control-plane
loadDockerImages:
  images:
    - stefanprodan/podinfo:latest
postInstallManifests:
  - "https://tinyurl.com/example-pod"
EOF
```

This configuration will:
- Create a KIND cluster named `mycluster`
- Install the Podinfo Helm chart with 2 replicas
- Pre-load the Podinfo image
- Deploy an example pod from a remote manifest

---

## Start the Cluster

Run BeKind with your configuration:

```bash
bekind start --config /tmp/config.yaml
```

BeKind will:
1. Create the KIND cluster
2. Load the specified Docker images
3. Install the Helm chart
4. Apply the remote manifest

This process typically takes 1-2 minutes.

---

## Verify the Deployment

Once BeKind completes, check that everything is running:

```bash
kubectl get pods -A
```

You should see your instance running with the Helm chart and pod deployed:

```
NAMESPACE            NAME                                              READY   STATUS    RESTARTS   AGE
default              example-pod-1                                     1/1     Running   0          6s
demo                 podinfo-d8689c8b7-qdjt4                           1/1     Running   0          24s
demo                 podinfo-d8689c8b7-v7r9d                           1/1     Running   0          24s
kube-system          coredns-66bc5c9577-2ndxz                          1/1     Running   0          24s
kube-system          coredns-66bc5c9577-8gfzx                          1/1     Running   0          24s
kube-system          etcd-mycluster-control-plane                      1/1     Running   0          31s
kube-system          kindnet-85crs                                     1/1     Running   0          24s
kube-system          kube-apiserver-mycluster-control-plane            1/1     Running   0          30s
kube-system          kube-controller-manager-mycluster-control-plane   1/1     Running   0          32s
kube-system          kube-proxy-dvxrq                                  1/1     Running   0          24s
kube-system          kube-scheduler-mycluster-control-plane            1/1     Running   0          30s
local-path-storage   local-path-provisioner-7b8c8ddbd6-8t47f           1/1     Running   0          24s
```

---

## Explore Your Cluster

### Check the Podinfo Application

```bash
kubectl get pods -n demo
kubectl describe pod -n demo -l app.kubernetes.io/name=podinfo
```

### Check the Example Pod

```bash
kubectl get pod example-pod-1
kubectl logs example-pod-1
```

### Access Podinfo (Port Forward)

```bash
kubectl port-forward -n demo svc/podinfo 9898:9898
```

Then visit [http://localhost:9898](http://localhost:9898) in your browser.

---

## Clean Up

When you're done, destroy the cluster:

```bash
bekind destroy --name mycluster
```

---

## Next Steps

Now that you've seen BeKind in action, explore more:

- [Installation Guide]({% link installation.md %}) - Different ways to install BeKind
- [CLI Commands]({% link cli-commands.md %}) - All available commands
- [Configuration Reference]({% link configuration.md %}) - All configuration options
- [Features]({% link features/index.md %}) - Deep dive into each feature
- [Profiles]({% link configuration.md %}#configuration-profiles) - Create reusable configurations

---

## Troubleshooting

### Docker Not Running

If you see errors about Docker:

```bash
# Check Docker is running
docker ps
```

### Cluster Already Exists

If a cluster named `mycluster` already exists:

```bash
# Destroy the existing cluster first
bekind destroy --name mycluster

# Or use a different name in your config.yaml
```

### Image Pull Issues

If images fail to pull:

```bash
# Pre-pull the image manually
docker pull stefanprodan/podinfo:latest

# Then run bekind again
bekind start --config /tmp/config.yaml
```
