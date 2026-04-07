# Getting Started

KubeMemo is a standalone Kubedeck tool for attaching durable and runtime context to Kubernetes resources.

## What you get

- Durable memos for warnings, ownership, maintenance guidance, runbooks, and curated operational memory.
- Runtime memos for incident context, temporary advice, and expiring breadcrumbs.
- A native Go CLI and TUI.
- A thin PowerShell wrapper module for PowerShell Gallery users.
- GitOps-aware behavior for durable note workflows.

## Quick start

Install KubeMemo from PowerShell Gallery:

```powershell
Install-Module -Name KubeMemo -Repository PSGallery -Scope CurrentUser
Import-Module KubeMemo
```

Bootstrap the cluster prerequisites:

```powershell
Install-KubeMemo -EnableRuntimeStore -RuntimeNamespace kubememo-runtime
```

Create a durable memo:

```powershell
New-KubeMemo `
  -Title "Orders API warm-up behavior" `
  -Summary "Expected transient 502s after deployment" `
  -Content "Ignore failures for up to 3 minutes after rollout." `
  -Kind Deployment `
  -Namespace prod `
  -Name orders-api `
  -NoteType warning
```

Read it back:

```powershell
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api
Open-KubeMemoTui
```

Start watching for captured activity:

```powershell
Start-KubeMemoActivityCapture -Namespace prod -Kind Deployment
```

## Core model

- Durable memos live in `memos.notes.kubememo.io`
- Runtime memos live in `runtimememos.runtime.notes.kubememo.io`
- Durable and runtime notes are intentionally separate stores

That split matters most in GitOps clusters, where durable objects may be Git-managed but runtime notes still need safe live writes.
