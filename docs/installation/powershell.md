# PowerShell Installation

KubeMemo is available as a PowerShell module for teams that prefer PowerShell Gallery installation and PowerShell-friendly command names.

## Install from PowerShell Gallery

```powershell
Install-Module -Name KubeMemo -Repository PSGallery -Scope CurrentUser
```

Import the module:

```powershell
Import-Module KubeMemo
```

## Bootstrap cluster prerequisites

Install the durable store:

```powershell
Install-KubeMemo
```

Install with runtime memos enabled:

```powershell
Install-KubeMemo -EnableRuntimeStore -RuntimeNamespace kubememo-runtime
```

Install with GitOps-aware checks:

```powershell
Install-KubeMemo -GitOpsAware -EnableRuntimeStore
```

Install with the optional in-cluster activity watcher:

```powershell
Install-KubeMemo `
  -EnableRuntimeStore `
  -EnableActivityCapture `
  -ActivityCaptureImage ghcr.io/kubedeckio/kubememo:0.0.1
```

## Typical PowerShell workflow

```powershell
New-KubeMemo -Title "Orders API warm-up behavior" -Summary "Expected transient 502s after deployment" -Content "Ignore failures for up to 3 minutes after rollout." -Kind Deployment -Namespace prod -Name orders-api -NoteType warning
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api
Open-KubeMemoTui
```

## How PowerShell fits the product

The PowerShell experience uses the native `kubememo` binary underneath, while keeping PowerShell-friendly command names, terminal passthrough for the TUI, and object-returning behavior where it fits PowerShell workflows.

## Related pages

- [Native CLI](native-cli.md)
- [Windows](windows.md)
- [Helm chart](helm.md)
- [Activity capture](../concepts/activity-capture.md)
