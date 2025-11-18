---
layout: default
title: Home
nav_order: 1
description: "BeKind - Install a K8S cluster using KIND with automated post-deployment steps"
permalink: /
---

# BeKind
{: .fs-9 }

Install a K8S cluster using [KIND](https://github.com/kubernetes-sigs/kind), and perform automated post-deployment steps.
{: .fs-6 .fw-300 }

[Quick Start]({% link quickstart.md %}){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View it on GitHub](https://github.com/christianh814/bekind){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## What is BeKind?

BeKind is a powerful CLI tool that simplifies the process of creating and configuring Kubernetes clusters using KIND (Kubernetes in Docker). It goes beyond basic cluster creation by automating common post-deployment tasks.

### Key Features

- **Automated Cluster Creation**: Deploy KIND clusters with customizable configurations
- **Custom Kubernetes Versions**: Specify any KIND node image version
- **Helm Chart Integration**: Automatically install Helm charts during cluster setup
- **Image Pre-loading**: Load Docker images into your cluster before deployment
- **Manifest Application**: Apply Kubernetes manifests automatically after cluster creation
- **Post-Install Actions**: Perform automated actions like resource restarts and deletions
- **Configuration Profiles**: Save and reuse cluster configurations

### Quick Example

```bash
# Start a cluster with the default configuration
bekind start

# Run a saved profile
bekind run argocd

# List all running clusters
bekind list

# Destroy a cluster
bekind destroy
```

---

## Why Use BeKind?

Setting up a Kubernetes development environment can be time-consuming, especially when you need to:
- Install multiple Helm charts
- Pre-load custom images
- Apply various manifests
- Restart deployments or clean up resources

BeKind automates all of these tasks through a simple YAML configuration file, allowing you to create reproducible Kubernetes environments in seconds.

Whether you're developing applications, testing Kubernetes features, or creating demo environments, BeKind streamlines your workflow.

---

## Documentation

### Getting Started
- [Quick Start]({% link quickstart.md %}) - Get up and running in minutes
- [Installation]({% link installation.md %}) - Install BeKind and prerequisites
- [CLI Commands]({% link cli-commands.md %}) - Complete command reference
- [Configuration]({% link configuration.md %}) - Configure your clusters

### Features
- [Helm Charts]({% link features/helm-charts.md %}) - Automatically install Helm charts
- [Loading Images]({% link features/loading-images.md %}) - Pre-load Docker images
- [Post Install Manifests]({% link features/post-install-manifests.md %}) - Apply Kubernetes manifests
- [Post Install Actions]({% link features/post-install-actions.md %}) - Automate resource operations

---

## Getting Help

If you have questions or run into issues:

- Check out the [CLI Commands]({% link cli-commands.md %}) reference
- Review the [Configuration Guide]({% link configuration.md %})
- Explore the [Features Documentation]({% link features/index.md %})
- Open an issue on [GitHub](https://github.com/christianh814/bekind/issues)

## License

BeKind is available as open source under the terms of the [Apache License 2.0](https://github.com/christianh814/bekind/blob/main/LICENSE).
