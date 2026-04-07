# Command Reference

## Bootstrap

- `Install-KubeMemo`
- `Uninstall-KubeMemo`
- `Update-KubeMemo`
- `Test-KubeMemoInstallation`
- `Get-KubeMemoInstallationStatus`

`Test-KubeMemoInstallation` and `Get-KubeMemoInstallationStatus` now include capability summaries so you can see whether the current identity can read durable memos, write runtime memos, patch annotations, and run always-on activity capture safely.

`Open-KubeMemoTui` also surfaces a compact access summary in the header and status line so namespace-scoped or read-only sessions are obvious while browsing.

## Read and search

- `Get-KubeMemo`
- `Find-KubeMemo`
- `Show-KubeMemo`
- `Open-KubeMemoTui`
- `Get-KubeMemoVersion`

## Write and remove

- `New-KubeMemo`
- `Set-KubeMemo`
- `Remove-KubeMemo`
- `Clear-KubeMemo`

## GitOps and config

- `Export-KubeMemo`
- `Import-KubeMemo`
- `Sync-KubeMemoGitOps`
- `Test-KubeMemoGitOpsMode`
- `Test-KubeMemoRuntimeStore`
- `Get-KubeMemoConfig`
- `Get-KubeMemoActivity`
- `Start-KubeMemoActivityCapture`

`Export-KubeMemo` now writes a clearer GitOps-oriented layout:

- `namespaces/<namespace>/...`
- `apps/<app>/...`
- `resources/<namespace>/<kind>-<name>/...`
- `runtime/<runtime-namespace>/...`

## Activity capture paths

KubeMemo supports three ways to run activity capture:

- foreground Go CLI watcher:
  - `kubememo watch-activity`
- PowerShell wrapper:
  - `Start-KubeMemoActivityCapture`
- always-on cluster deployment:
  - [Helm chart](../installation/helm.md)

## Common examples

```powershell
Get-KubeMemo -Namespace prod
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api
New-KubeMemo -Temporary -Title "Investigation" -Summary "Watching rollout" -Content "Temporary runtime note" -Kind Deployment -Namespace prod -Name orders-api -NoteType incident
```
