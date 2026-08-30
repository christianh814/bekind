---
layout: default
title: Variables
parent: Features
nav_order: 7
description: "Define reusable variables in your BeKind configuration"
---

# Variables
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

BeKind lets you define key/value pair variables at the top of your configuration file and reference them anywhere else in the YAML, including Helm chart values, Docker image lists, the KIND config, manifests lists, and Helm Stack files. This is useful for:

- Defining a value once (like a domain or image tag) and reusing it in multiple places
- Keeping configurations easy to update
- Building values from other values (like a domain built from an IP)

---

## Configuration

Declare variables under the top-level `vars` key as a list of `name`/`value` pairs, then reference them with `${{ .vars.<name> }}`:

```yaml
vars:
  - name: myimage
    value: quay.io/christianh814/simple-go:latest
loadDockerImages:
  images:
    - ${{ .vars.myimage }}
```

Variables work anywhere in the configuration, including nested Helm values:

```yaml
vars:
  - name: domainName
    value: 7f000001.nip.io
helmCharts:
  - url: "https://example.com/charts"
    repo: "example"
    chart: "test"
    release: "test"
    namespace: "mynamespace"
    wait: true
    valuesObject:
      installCRDs: "true"
      controller:
        domains:
          - ${{ .vars.domainName }}
```

---

## Configuration Options

### name

**Type**: `string`  
**Required**: Yes  
**Description**: The name of the variable. Must start with a letter or underscore and contain only letters, numbers, and underscores.

Names that match a BeKind configuration field (for example `loadDockerImages`, `kindConfig`, or `helmCharts`) are reserved and cannot be used. Duplicate names are also rejected.

### value

**Type**: `string`  
**Required**: Yes  
**Description**: The value of the variable. A value may reference variables defined **earlier** in the list:

```yaml
vars:
  - name: ip
    value: 7f000001
  - name: domainName
    value: ${{ .vars.ip }}.nip.io
```

---

## Variables in Helm Stacks

Variables defined in your main configuration file are also expanded inside [Helm Stack]({% link features/helm-charts.md %}) `stack.yaml` files:

```yaml
# ~/.bekind/config.yaml
vars:
  - name: domainName
    value: 7f000001.nip.io
helmStack:
  - name: argocd
```

```yaml
# ~/.bekind/helmstack/argocd/stack.yaml
helmCharts:
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: true
    valuesObject:
      global:
        domain: argocd.${{ .vars.domainName }}
```

---

## Error Handling

BeKind fails fast with a clear error if:

- A variable name is invalid, reserved, or duplicated
- A value references a variable that hasn't been defined yet
- The configuration references an undefined variable (catches typos)

{: .note }
Referencing an undefined variable is always an error, so a typo like `${{ .vars.myimge }}` will stop cluster creation instead of being silently ignored.

---

## Next Steps

- [Configuration Reference]({% link configuration.md %})
- [Helm Charts]({% link features/helm-charts.md %})
- [Loading Docker Images]({% link features/loading-images.md %})
