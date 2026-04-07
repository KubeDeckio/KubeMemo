# Installation

## PowerShell Gallery

```powershell
Install-Module -Name KubeMemo -Repository PSGallery -Scope CurrentUser
```

## Cluster bootstrap

Install the CRDs:

```powershell
Install-KubeMemo
```

Install with runtime store support:

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

## Build the Go CLI

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
- For docs development: MkDocs Material
