# Command Reference

## Bootstrap

- `Install-KubeMemo`
- `Uninstall-KubeMemo`
- `Update-KubeMemo`
- `Test-KubeMemoInstallation`
- `Get-KubeMemoInstallationStatus`

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

## Common examples

```powershell
Get-KubeMemo -Namespace prod
Show-KubeMemo -Kind Deployment -Namespace prod -Name orders-api
New-KubeMemo -Temporary -Title "Investigation" -Summary "Watching rollout" -Content "Temporary runtime note" -Kind Deployment -Namespace prod -Name orders-api -NoteType incident
```
