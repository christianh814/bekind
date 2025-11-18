---
layout: default
title: Installation
nav_order: 3
description: "How to install BeKind"
---

# Installation
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Prerequisites

Before installing BeKind, ensure you have the following installed:

### Docker

BeKind uses KIND, which requires Docker to run Kubernetes clusters in containers.

{: .note }
Podman is still [considered experimental](https://kind.sigs.k8s.io/docs/design/principles/#target-cri-functionality) by KIND and may not work reliably with BeKind.

To check if Docker is installed:

```bash
docker --version
```

To install Docker:
- **Linux**: Follow the [Docker Engine installation guide](https://docs.docker.com/engine/install/)
- **macOS**: Install [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/)
- **Windows**: Install [Docker Desktop for Windows](https://docs.docker.com/desktop/install/windows-install/)

Ensure Docker is running before using BeKind:

```bash
docker ps
```

---

## Installing BeKind

### Download from Releases Page

1. Visit the [BeKind Releases page](https://github.com/christianh814/bekind/releases)

2. Download the latest release for your platform:
   - **Linux (AMD64)**: `bekind-linux-amd64`
   - **macOS (ARM64/Apple Silicon)**: `bekind-darwin-arm64`

3. Make the binary executable and move it to your PATH:

**Linux:**
```bash
chmod +x bekind-linux-amd64
sudo mv bekind-linux-amd64 /usr/local/bin/bekind
```

**macOS:**
```bash
chmod +x bekind-darwin-arm64
sudo mv bekind-darwin-arm64 /usr/local/bin/bekind
```

{: .note }
On macOS, you may need to allow the binary to run by going to System Preferences → Security & Privacy if you see a security warning on first run.

---

## Verifying Installation

To verify BeKind is installed correctly:

```bash
bekind version
```

You should see output showing the BeKind version information.

To see all available commands:

```bash
bekind --help
```

---

## Next Steps

Now that BeKind is installed, you can:

1. [Review CLI commands]({% link cli-commands.md %}) - Learn all available commands
2. [Learn about configuration options]({% link configuration.md %}) - Customize your clusters
3. [Explore features]({% link features/index.md %}) - Discover what BeKind can do
4. Start your first cluster with `bekind start`
