# Installation

KubeMemo has three supported installation paths:

- Native CLI from Homebrew or GitHub Releases
- PowerShell module from PowerShell Gallery
- Optional Helm deployment for in-cluster activity capture

The native CLI is the core product experience. If you work primarily in PowerShell, KubeMemo is also available as a PowerShell module with PowerShell-friendly commands.

Use the path that best matches how your team works:

- [Native CLI installation](installation/native-cli.md)
- [Windows installation](installation/windows.md)
- [PowerShell installation](installation/powershell.md)
- [Helm chart deployment](installation/helm.md)

The Helm chart is published to GHCR as an OCI artifact:

```bash
helm install kubememo oci://ghcr.io/kubedeckio/charts/kubememo --version 0.0.1
```

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

Enable the optional in-cluster activity watcher during install:

```powershell
Install-KubeMemo -EnableRuntimeStore -EnableActivityCapture -ActivityCaptureImage ghcr.io/kubedeckio/kubememo:0.0.1
```

Check cluster prerequisites and current RBAC capabilities:

```powershell
Get-KubeMemoInstallationStatus
Test-KubeMemoInstallation
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

## RBAC expectations

KubeMemo works with different permission levels depending on how you want to use it.

### Read-only use

For read-only usage, the current identity needs:

- `get`, `list`, `watch` on `memos.notes.kubememo.io`
- `get`, `list`, `watch` on `runtimememos.runtime.notes.kubememo.io`

This is enough for:

- `Get-KubeMemo`
- `Find-KubeMemo`
- `Show-KubeMemo`
- `Open-KubeMemoTui`

### Memo write use

For creating and updating memos, the current identity also needs:

- `create`, `update`, `patch`, `delete` on `memos.notes.kubememo.io`
- `create`, `update`, `patch`, `delete` on `runtimememos.runtime.notes.kubememo.io`

If you want annotation sync on target resources, the identity also needs:

- `patch` on the target resource kinds you want KubeMemo to annotate

KubeMemo now uses lightweight annotation patching rather than broader full-object updates.

### Install and admin use

For `Install-KubeMemo`, `Update-KubeMemo`, and the optional in-cluster watcher deployment, the current identity may also need:

- CRD management permissions
- namespace create/update permissions
- RBAC object create/update permissions
- Deployment create/update permissions in the runtime namespace

### Optional in-cluster activity capture

The always-on watcher needs:

- `get`, `list`, `watch` on watched resource kinds such as Deployments, Services, Ingresses, HPAs, Gateways, and HTTPRoutes
- `create`, `update`, `patch`, `get`, `list` on runtime memos

Use `Get-KubeMemoInstallationStatus` or `Test-KubeMemoInstallation` to see the current capability summary for the active identity.
