<p align="center">
  <img src="./images/KubeMemo.png" />
</p>
<h1 align="center" style="font-size: 100px;">
  <b>KubeMemo</b>
</h1>

</br>

![PowerShell Gallery Version](https://img.shields.io/powershellgallery/v/KubeMemo.svg)
![Downloads](https://img.shields.io/powershellgallery/dt/KubeMemo.svg)
![License](https://img.shields.io/github/license/KubeDeckio/KubeMemo.svg)

---

**KubeMemo** is a PowerShell-first Kubernetes operational memory tool from **Kubedeck**. It lets you attach durable notes, temporary runtime notes, and activity breadcrumbs to Kubernetes resources, namespaces, and logical applications so operational context stays with the workload.

## Documentation

KubeMemo is designed as a standalone Kubedeck tool with future integration points for KubeBuddy.

## Features

- **Durable Notes:** Store long-lived runbooks, ownership notes, warnings, and maintenance guidance in a dedicated CRD.
- **Runtime Notes:** Capture temporary handover notes, incident context, and expiring operational breadcrumbs in a separate runtime CRD.
- **GitOps-Aware Behavior:** Block direct durable writes in GitOps mode and generate file-based output for Git-managed workflows.
- **Cluster Bootstrap:** Install and validate CRDs, runtime namespace, and optional RBAC directly from the PowerShell module.
- **Memo-Style Rendering:** Show durable and runtime context as memo-style terminal cards with color, wrapped content, and clearer note sections.
- **Built-In TUI:** Browse and view notes from an interactive memo board with `Open-KubeMemoTui`.
- **RBAC-Aware Reads:** Fall back to namespace-scoped reads when cluster-wide listing is not allowed, and filter the TUI by namespace scope.
- **Actor Stamping:** Record `CreatedBy` and `UpdatedBy` for memos, preferring the RBAC identity from `kubectl auth whoami` when available.
- **Unified Object Model:** Normalize durable and runtime notes into one PowerShell shape for search, filtering, and rendering.
- **Runtime Cleanup:** Remove expired runtime notes without touching durable note data.

## Resource Model

KubeMemo uses separate durable and runtime stores.

- **Durable CRD:** `memos.notes.kubememo.io`
- **Runtime CRD:** `runtimememos.runtime.notes.kubememo.io`
- **Short names:** `km`, `memo`, `kmr`, `rmemo`

These shorter resource names are the intended v1 direction. On an existing test cluster, moving from older CRD names to the shorter names requires a clean reinstall because CRD identity changes are breaking.

## Installation

### PowerShell Gallery

To install **KubeMemo** via PowerShell Gallery:

```powershell
Install-Module -Name KubeMemo -Repository PSGallery -Scope CurrentUser
```

### Bootstrap Cluster Prerequisites

Install the KubeMemo CRDs:

```powershell
Install-KubeMemo
```

Install KubeMemo with the runtime store enabled:

```powershell
Install-KubeMemo -EnableRuntimeStore -RuntimeNamespace kubememo-runtime
```

Install KubeMemo in GitOps-aware mode:

```powershell
Install-KubeMemo -GitOpsAware -EnableRuntimeStore
```

## Usage

Create a durable note:

```powershell
New-KubeMemo -Title "Orders API warm-up behavior" -Summary "Expected transient 502s after deployment" -Content "Ignore failures for up to 3 minutes after rollout." -Kind Deployment -Namespace prod -Name orders-api -NoteType advisory
```

Create a durable note with explicit ownership metadata:

```powershell
New-KubeMemo -Title "Payments API owner note" -Summary "Service owned by platform-apps" -Content "Contact @platform-apps before changing ingress or HPA settings." -Kind Deployment -Namespace prod -Name orders-api -NoteType ownership -OwnerTeam "platform-apps" -OwnerContact "@platform-apps"
```

Create a temporary runtime note:

```powershell
New-KubeMemo -Temporary -Title "Investigating replica spike" -Summary "Manual scale change under review" -Content "Replicas were increased during the incident call." -Kind Deployment -Namespace prod -Name orders-api -NoteType incident -ExpiresAt (Get-Date).ToUniversalTime().AddHours(12)
```

Show notes for a resource:

```powershell
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api -IncludeRuntime
```

Show notes with plain output and no ANSI colors:

```powershell
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api -IncludeRuntime -NoColor
```

Open the interactive memo board:

```powershell
Open-KubeMemoTui -IncludeRuntime
```

Open the memo board scoped to one namespace:

```powershell
Open-KubeMemoTui -IncludeRuntime -Namespace prod
```

TUI shortcuts:

```text
[Arrows]/[j][k] move
[PgUp]/[PgDn] or [u][d] scroll the detail pane
[/] text filter
[:] switch view (:memo, :runtimememo, :all)
[f] namespace filter
[c] kind filter
[a] add a temporary runtime memo
[r] refresh
[q] quit
```

Show creator/updater metadata:

```powershell
Get-KubeMemo -IncludeRuntime | Select-Object Id, StoreType, CreatedBy, UpdatedBy, Title | Format-Table -AutoSize
```

When available, KubeMemo uses `kubectl auth whoami` so the actor reflects the RBAC username seen by the cluster, not just the local shell user.

Export durable notes for GitOps workflows:

```powershell
Export-KubeMemo -Path ./ops/kubememo
```

## Command Reference

Public commands exported by the module:

### Bootstrap and lifecycle

- `Install-KubeMemo` installs CRDs, runtime namespace, and optional RBAC.
- `Uninstall-KubeMemo` removes KubeMemo prerequisites, with runtime-only and data-preservation options.
- `Update-KubeMemo` reapplies bundled manifests and updates installed prerequisites.
- `Test-KubeMemoInstallation` checks cluster reachability, CRDs, namespace, RBAC, and GitOps/runtime state.
- `Get-KubeMemoInstallationStatus` returns the detected installation mode and status summary.

### Read, search, and render

- `Get-KubeMemo` returns normalized durable notes, or durable plus runtime notes with `-IncludeRuntime`.
- `Find-KubeMemo` filters notes by text, type, kind, namespace, name, and expiry handling.
- `Show-KubeMemo` renders memo-style cards for a target in the terminal.
- `Open-KubeMemoTui` opens the interactive memo board for browsing and reading notes.

### Write and remove

- `New-KubeMemo` creates a durable memo by default, or a runtime memo with `-Temporary`.
- `Set-KubeMemo` edits an existing memo.
- `Remove-KubeMemo` removes a memo or deletes expired runtime memos.
- `Clear-KubeMemo` clears runtime memos, with `-ExpiredOnly` support.

### GitOps and portability

- `Export-KubeMemo` writes memo manifests to files for Git-managed workflows.
- `Import-KubeMemo` applies memo manifests from disk.
- `Sync-KubeMemoGitOps` runs GitOps-oriented import or export flows.
- `Test-KubeMemoGitOpsMode` reports the detected GitOps state.
- `Test-KubeMemoRuntimeStore` validates runtime-store availability and safety.

### Diagnostics and activity

- `Get-KubeMemoActivity` returns runtime activity memos.
- `Get-KubeMemoConfig` returns the effective internal configuration.

Private helper functions under [KubeMemo/Private](/Users/pixelrobots/Documents/Git/KubeMemo/KubeMemo/Private) are internal implementation details and are not part of the supported public interface.

## Changelog

All notable changes to this project should be documented in the [CHANGELOG](./CHANGELOG.md).

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for more details.
