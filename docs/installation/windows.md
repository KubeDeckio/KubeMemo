# Windows

KubeMemo supports Windows through:

- the native `kubememo.exe` release asset
- the PowerShell wrapper module from PowerShell Gallery

## Native CLI on Windows

1. Download the Windows release archive from GitHub Releases.
2. Extract `kubememo.exe`.
3. Place it in a directory on your `PATH`.

Verify:

```powershell
kubememo.exe version --output json
kubememo.exe status --output json
```

## PowerShell on Windows

If you prefer PowerShell-native commands, install the module:

```powershell
Install-Module -Name KubeMemo -Repository PSGallery -Scope CurrentUser
```

Then bootstrap the cluster:

```powershell
Install-KubeMemo -EnableRuntimeStore -InstallRbac
```

## Activity capture

Windows users can use either:

- the foreground watcher:

```powershell
Start-KubeMemoActivityCapture -Namespace prod -Kind Deployment
```

- or the optional in-cluster deployment:

- [Helm chart deployment](helm.md)
