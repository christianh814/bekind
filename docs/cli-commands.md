---
layout: default
title: CLI Commands
nav_order: 5
description: "BeKind CLI commands reference"
---

# CLI Commands
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

BeKind provides several commands to manage your KIND clusters. This page documents all available commands and their options.

---

## bekind start

Start a new KIND cluster with the specified configuration.

### Usage

```bash
bekind start [flags]
```

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--config` | string | Path to config file | `$HOME/.bekind/config.yaml` |
| `--name` | string | Name of the KIND cluster | `kind` |

### Examples

**Start with default configuration:**
```bash
bekind start
```

**Start with a custom config file:**
```bash
bekind start --config /path/to/config.yaml
```

**Start with a custom cluster name:**
```bash
bekind start --name my-cluster
```

**Combine flags:**
```bash
bekind start --config /path/to/config.yaml --name dev-cluster
```

### Behavior

When you run `bekind start`:

1. Reads the configuration file (default: `~/.bekind/config.yaml`)
2. Creates a KIND cluster with the specified settings
3. Loads Docker images (if configured)
4. Installs Helm charts (if configured)
5. Applies Kubernetes manifests (if configured)
6. Performs post-install actions (if configured)

---

## bekind run

Run a saved profile. Profiles allow you to save and reuse cluster configurations.

### Usage

```bash
bekind run <profile-name> [flags]
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `profile-name` | Yes | Name of the profile to run |

### Flags

| Flag | Short | Type | Description | Default |
|------|-------|------|-------------|---------|
| `--view` | `-v` | boolean | View the profile configuration without running it | `false` |
| `--profile-dir` | `-p` | string | Directory where profiles are stored | `$HOME/.bekind/profiles` |
| `--name` | | string | Name of the KIND cluster | `kind` |

### Examples

**Run a profile:**
```bash
bekind run argocd
```

This looks for configuration files in `~/.bekind/profiles/argocd/`.

**View profile configuration without running:**
```bash
bekind run argocd --view
```

**Run profile from custom directory:**
```bash
bekind run myprofile --profile-dir /tmp
```

This looks for configuration files in `/tmp/myprofile/`.

**Run profile with custom cluster name:**
```bash
bekind run argocd --name my-cluster
```

### Profile Structure

Profiles are stored in directories under `~/.bekind/profiles/`:

```
~/.bekind/profiles/
├── argocd/
│   ├── config.yaml
│   └── extra-config.yaml
├── dev/
│   └── config.yaml
└── production/
    └── config.yaml
```

### Multiple Configuration Files

If a profile directory contains multiple `.yaml` files, `bekind run` will execute each configuration file in sequence:

```bash
# Given this structure:
~/.bekind/profiles/argocd/
├── base-cluster.yaml
├── argocd-setup.yaml
└── apps.yaml

# Running this:
bekind run argocd

# Will execute all three YAML files
```

### Behavior

When you run `bekind run <profile>`:

1. Locates the profile directory (`~/.bekind/profiles/<profile>/`)
2. Finds all `.yaml` files in the directory
3. For each configuration file:
   - Creates/updates the KIND cluster
   - Applies the configuration
   - Resets state between files
4. Multiple configs in the same profile can create different clusters or update the same cluster

---

## bekind list

List all running KIND clusters.

### Usage

```bash
bekind list
```

### Aliases

`ls`

### Examples

**List all clusters:**
```bash
bekind list
# or
bekind ls
```

**Example output:**
```
kind
argocd-cluster
dev-cluster
```

If no clusters are found:
```
INFO[0000] No clusters found
```

### Behavior

Lists all KIND clusters on the system, including:
- Clusters created by BeKind
- Clusters created directly with KIND
- Any cluster managed by KIND, regardless of origin

---

## bekind destroy

Destroy (delete) a KIND cluster.

### Usage

```bash
bekind destroy [flags]
```

### Aliases

`delete`, `del`

### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--name` | string | Name of the cluster to destroy | `kind` |
| `--config` | string | Config file to read cluster name from | `$HOME/.bekind/config.yaml` |

### Examples

**Destroy the default cluster:**
```bash
bekind destroy
```

**Destroy a specific cluster by name:**
```bash
bekind destroy --name argocd-cluster
```

**Use aliases:**
```bash
bekind delete --name my-cluster
bekind del --name my-cluster
```

**Destroy cluster specified in config:**
```bash
bekind destroy --config /path/to/config.yaml
```

### Behavior

The `destroy` command will:
1. Check if a cluster name is specified with `--name`
2. If a config file is provided, read the cluster name from the `kindConfig.name` field
3. Delete the specified KIND cluster

{: .note }
If the cluster name is specified in both the `--name` flag and the config file, the config file takes precedence.

---

## bekind purge

Remove all KIND clusters.

### Usage

```bash
bekind purge [flags]
```

### Flags

| Flag | Short | Type | Description | Default |
|------|-------|------|-------------|---------|
| `--confirm` | `-c` | boolean | Skip confirmation prompt and purge immediately | `false` |

### Examples

**Purge with confirmation prompt:**
```bash
bekind purge
# Are you sure you want to delete all KIND clusters on the system? [y/N]:
```

**Purge without confirmation (bypass prompt):**
```bash
bekind purge --confirm
# or
bekind purge -c
```

{: .warning }
This command destroys **all** KIND clusters on your system, regardless of whether they were created by BeKind or not. Use with caution!

---

## bekind showconfig

Display the current configuration.

### Usage

```bash
bekind showconfig [flags]
```

### Aliases

`sc`, `configShow`

### Flags

| Flag | Short | Type | Description | Default |
|------|-------|------|-------------|---------|
| `--system` | `-s` | boolean | Print the config saved on the cluster | `false` |
| `--config` | | string | Config file to display | `$HOME/.bekind/config.yaml` |

### Examples

**Show default configuration:**
```bash
bekind showconfig
```

**Show specific config file:**
```bash
bekind showconfig --config /path/to/config.yaml
```

**Show configuration saved on the cluster:**
```bash
bekind showconfig --system
# or
bekind showconfig -s
```

**Use aliases:**
```bash
bekind sc
bekind configShow
```

### Behavior

The `showconfig` command displays configuration in two modes:

**Local mode (default):**
- Reads the configuration file from disk
- Shows the merged configuration from Viper
- Displays what would be used when starting a cluster

**System mode (`--system` flag):**
- Connects to a running cluster
- Retrieves the configuration from the `bekind-config` secret in the `kube-public` namespace
- Shows the configuration that was used when the cluster was created

This is useful for:
- Verifying your configuration before starting a cluster
- Debugging configuration issues
- Comparing local config with what's deployed
- Documenting cluster setup

---

## bekind version

Display the BeKind version.

### Usage

```bash
bekind version
```

### Examples

```bash
bekind version
```

**Example output:**
```json
{"bekind":"v0.5.1"}
```

### Behavior

Returns version information in JSON format with the command name as the key and version as the value.

---

## Global Flags

These flags are available for all commands:

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--config` | string | Config file path | `$HOME/.bekind/config.yaml` |
| `--name` | string | KIND cluster name | `kind` |

---

## Tab Completion

BeKind supports shell completion for commands, flags, and profile names.

### Bash

```bash
bekind completion bash > /etc/bash_completion.d/bekind
```

### Zsh

```bash
bekind completion zsh > "${fpath[1]}/_bekind"
```

### Fish

```bash
bekind completion fish > ~/.config/fish/completions/bekind.fish
```

### PowerShell

```powershell
bekind completion powershell > bekind.ps1
```

---

## Common Workflows

### Quick Start Development Cluster

```bash
bekind start
```

### Use a Saved Profile

```bash
bekind run argocd
```

### Check What's Configured

```bash
bekind run argocd --view
```

### Multiple Environments

```bash
# Development
bekind run dev --name dev-cluster

# Staging
bekind run staging --name staging-cluster

# Clean up
bekind destroy dev-cluster
bekind destroy staging-cluster
```

### Test Configuration Changes

```bash
# View the config first
bekind showconfig --config ./new-config.yaml

# If it looks good, start the cluster
bekind start --config ./new-config.yaml
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error occurred |

---

## Tips

### Profile vs Config File

- **Use `bekind run <profile>`** when you have multiple reusable configurations
- **Use `bekind start --config`** for one-off configurations or testing

### Multiple Clusters

You can run multiple KIND clusters simultaneously by using different `--name` values:

```bash
bekind start --name cluster1
bekind start --name cluster2 --config other-config.yaml
bekind list
# cluster1
# cluster2
```

### Debugging

Use `bekind showconfig` to verify your configuration before starting a cluster:

```bash
bekind showconfig --config ./my-config.yaml
```

### Cleaning Up

Remove all clusters when done:

```bash
bekind purge
```

Or remove specific clusters:

```bash
bekind destroy cluster1
bekind destroy cluster2
```
