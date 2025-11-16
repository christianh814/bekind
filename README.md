# BeKind

A powerful CLI tool for creating and configuring Kubernetes clusters using KIND (Kubernetes in Docker) with automated post-deployment steps.

## Documentation

For complete documentation, including installation instructions, configuration options, and feature guides, please visit:

**[https://christianh814.github.io/bekind/](https://christianh814.github.io/bekind/)**

## Quick Start

```bash
# Install
go install github.com/christianh814/bekind@latest

# Start a cluster with default config
bekind start

# Or run a saved profile
bekind run myprofile

# List clusters
bekind list

# Destroy a cluster
bekind destroy
```

## Features

- **Automated Cluster Creation**: Deploy KIND clusters with customizable configurations
- **Helm Chart Integration**: Automatically install Helm charts during cluster setup
- **Image Pre-loading**: Load Docker images into your cluster before deployment
- **Manifest Application**: Apply Kubernetes manifests automatically after cluster creation
- **Post-Install Actions**: Perform automated actions like resource restarts and deletions

## License

Distributed under the Apache License 2.0. See `LICENSE` for more information.
