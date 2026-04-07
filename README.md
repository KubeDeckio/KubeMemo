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

**KubeMemo** is a Kubernetes operational memory tool from **Kubedeck**. The core implementation is now a native **Go CLI and TUI**, with a thin **PowerShell wrapper module** so the tool can still be distributed through PowerShell Gallery.

## Documentation

KubeMemo is designed as a standalone Kubedeck tool with future integration points for KubeBuddy.

- Docs site: https://kubememo.kubedeck.io
- Source: `/docs`, `/overrides`, and `mkdocs.yml`

Build the docs locally:

```bash
python3 -m pip install -r requirements-docs.txt
mkdocs serve
```

## Release Automation

GitHub Actions now handle:

- tag-driven Go binary release assets
- tag-driven PowerShell Gallery publishing
- MkDocs deployment to GitHub Pages

Required repository secrets:

- `PSGALLERY_API_KEY` for PowerShell Gallery publishing

Tagging a release like `v0.1.1` triggers:

- cross-platform Go binary builds
- a GitHub release with release assets
- a PowerShell module package containing the wrapper and bundled binaries

## Architecture

- **Go core:** the real product surface lives in the `kubememo` Go binary.
- **PowerShell wrapper:** the PowerShell module shells out to the Go binary, parses JSON for object-returning commands, and keeps `WhatIf` / `Confirm` behavior for PowerShell users.
- **Cross-platform distribution:** the Go binary is the path for GitHub Releases, Homebrew, and future Krew packaging.

## Features

- **Durable Notes:** Store long-lived runbooks, ownership notes, warnings, and maintenance guidance in a dedicated CRD.
- **Runtime Notes:** Capture temporary handover notes, incident context, and expiring operational breadcrumbs in a separate runtime CRD.
- **GitOps-Aware Behavior:** Block direct durable writes in GitOps mode and generate file-based output for Git-managed workflows.
- **Cluster Bootstrap:** Install and validate CRDs, runtime namespace, and optional RBAC directly from the PowerShell module.
- **Memo-Style Rendering:** Show durable and runtime context as memo-style terminal cards with color, wrapped content, and clearer note sections.
- **Built-In TUI:** Browse and view notes from an interactive memo board with `kubememo tui` or `Open-KubeMemoTui`.
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

Update installed prerequisites with GitOps-aware checks:

```powershell
Update-KubeMemo -GitOpsAware
```

### Go CLI Build

Build the native CLI locally:

```bash
make build
```

Run the native CLI directly:

```bash
./KubeMemo/bin/$(go env GOOS)-$(go env GOARCH)/kubememo get --output json
```

## Usage

Native CLI:

```bash
kubememo get --namespace prod --output json
kubememo show --kind Deployment --namespace prod --name orders-api
kubememo tui
```

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
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api
```

Show notes with plain output and no ANSI colors:

```powershell
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api -NoColor
```

Open the interactive memo board:

```powershell
Open-KubeMemoTui
```

Open the memo board scoped to one namespace:

```powershell
Open-KubeMemoTui -Namespace prod
```

Force durable-only output for human-facing commands:

```powershell
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api -DurableOnly
Find-KubeMemo -Text rollout -DurableOnly
Open-KubeMemoTui -DurableOnly
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

Write a durable memo manifest to disk instead of applying it directly:

```powershell
New-KubeMemo -Title "Orders API deploy note" -Summary "Git-managed memo" -Content "Create this as a manifest for GitOps." -Kind Deployment -Namespace prod -Name orders-api -OutputPath ./ops/kubememo/resources/prod/deployment-orders-api.yaml
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

- `Get-KubeMemo` returns normalized durable notes by default, or durable plus runtime notes with `-IncludeRuntime`.
- `Find-KubeMemo` searches durable and runtime notes by default, with `-DurableOnly` available from the PowerShell wrapper.
- `Show-KubeMemo` renders memo-style cards for a target in the terminal and includes runtime notes by default.
- `Open-KubeMemoTui` opens the interactive memo board for browsing and reading notes, including runtime notes by default.

### Write and remove

- `New-KubeMemo` creates a durable memo by default, or a runtime memo with `-Temporary`.
- `Set-KubeMemo` edits an existing memo.
- `Remove-KubeMemo` removes a memo or deletes expired runtime memos.
- `Clear-KubeMemo` clears runtime memos, with `-ExpiredOnly` support.
- `New-KubeMemo` and `Set-KubeMemo` support `-OutputPath` for Git-managed durable manifests.
- `-AnnotateResource` remains unsupported in the Go core today and will fail clearly if used from the PowerShell wrapper.

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

## Development

Format the Go code:

```bash
make fmt
```

Run the Go test suite:

```bash
make test
```

Release builds are configured in [.goreleaser.yaml](/Users/pixelrobots/Documents/Git/KubeMemo/.goreleaser.yaml).

## Changelog

All notable changes to this project should be documented in the [CHANGELOG](./CHANGELOG.md).

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for more details.
