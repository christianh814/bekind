---
layout: default
title: Features
nav_order: 6
has_children: true
description: "BeKind features and capabilities"
---

# Features

BeKind provides several powerful features to automate your Kubernetes cluster setup and configuration.

## Available Features

### [Helm Charts]({% link features/helm-charts.md %})
Automatically install Helm charts during cluster creation with custom values and configurations.

### [Loading Docker Images]({% link features/loading-images.md %})
Pre-load Docker images into your KIND cluster nodes before starting your applications.

### [Post Install Manifests]({% link features/post-install-manifests.md %})
Apply Kubernetes YAML manifests automatically after cluster setup is complete.

### [Post Install Actions]({% link features/post-install-actions.md %})
Perform automated actions on Kubernetes resources, such as restarting deployments or deleting pods.

### [Post Install Patches]({% link features/post-install-patches.md %})
Apply JSON Patch operations to Kubernetes resources for fine-grained modifications after installation.

---

Each feature can be configured independently in your BeKind configuration file. You can use one, some, or all features depending on your needs.
