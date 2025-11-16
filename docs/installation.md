---
layout: default
title: Installation
nav_order: 2
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

### Go

BeKind requires Go version `1.20` or newer.

To check your Go version:

```bash
go version
```

If you need to install or update Go, visit the [official Go downloads page](https://golang.org/dl/).

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

### Using Go Install

The easiest way to install BeKind is using `go install`:

```bash
go install github.com/christianh814/bekind@latest
```

This will download and compile BeKind, placing the binary in your `$GOBIN` directory (typically `$GOPATH/bin` or `$HOME/go/bin`).

### Moving to Your PATH

After installation, move the binary to a directory in your `$PATH` for easy access. For example, to move it to `/usr/local/bin`:

```bash
sudo mv $GOBIN/bekind /usr/local/bin/bekind
sudo chmod +x /usr/local/bin/bekind
```

Alternatively, you can add `$GOBIN` to your `$PATH`:

```bash
# Add to your .bashrc, .zshrc, or equivalent
export PATH="$PATH:$(go env GOBIN)"
```

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
