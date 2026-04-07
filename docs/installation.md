# Installation

KubeMemo has three supported installation paths:

- Native CLI from Homebrew or GitHub Releases
- PowerShell module from PowerShell Gallery
- Optional Helm deployment for in-cluster activity capture

The native CLI is the core product experience. If you work primarily in PowerShell, KubeMemo is also available as a PowerShell module with PowerShell-friendly commands.

Use the path that best matches how your team works:

- [Native CLI installation](installation/macos-linux.md)
- [PowerShell installation](installation/powershell.md)
- [Helm activity-capture deployment](installation/macos-linux.md#optional-in-cluster-activity-capture-with-helm)

## Native CLI

Install with Homebrew or download a release asset, then verify the binary:

```bash
kubememo version --output json
kubememo status --output json
```

## PowerShell

Install from PowerShell Gallery:

```powershell
Install-Module -Name KubeMemo -Repository PSGallery -Scope CurrentUser
```

## Cluster bootstrap

Whichever install path you choose, the next step is bootstrapping the cluster prerequisites:

```powershell
Install-KubeMemo -EnableRuntimeStore -RuntimeNamespace kubememo-runtime
```

Install in GitOps-aware mode:

```powershell
Install-KubeMemo -GitOpsAware -EnableRuntimeStore
```

Update installed prerequisites:

```powershell
Update-KubeMemo -GitOpsAware
```

## Build locally

```bash
make build
```

Run the binary directly:

```bash
./KubeMemo/bin/$(go env GOOS)-$(go env GOARCH)/kubememo status --output json
```

## Prerequisites

- Access to a Kubernetes cluster
- `kubectl` configured for that cluster
- For PowerShell workflows: PowerShell 7+
- For Helm workflows: Helm 3+
- For docs development: MkDocs Material
