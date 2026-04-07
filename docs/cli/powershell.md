# PowerShell

KubeMemo also ships as a PowerShell module for users who want PowerShell-native commands on top of the core CLI.

## What the PowerShell experience does

- Finds the bundled binary for the current OS and architecture
- Passes through terminal commands for color and TUI support
- Parses JSON results for object-returning commands
- Preserves PowerShell-friendly command names and `WhatIf`/`Confirm` behavior

## Examples

```powershell
Get-KubeMemo -Namespace prod
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api
Open-KubeMemoTui
Get-KubeMemoVersion
Start-KubeMemoActivityCapture -Namespace prod -Kind Deployment
```

## Actor stamping

When possible, KubeMemo uses:

```bash
kubectl auth whoami -o json
```

That means `CreatedBy` and `UpdatedBy` reflect the RBAC-facing cluster identity rather than only the local shell user.
