function Get-KubeMemoBinaryPath {
    [CmdletBinding()]
    param()

    $moduleRoot = Split-Path -Parent $PSScriptRoot
    $repoRoot = Split-Path -Parent $moduleRoot

    $osName = if ($IsMacOS) { 'darwin' } elseif ($IsLinux) { 'linux' } else { 'windows' }
    $archName = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()) {
        'x64' { 'amd64' }
        'arm64' { 'arm64' }
        default { [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant() }
    }
    $binaryName = if ($osName -eq 'windows') { 'kubememo.exe' } else { 'kubememo' }
    $candidate = Join-Path -Path $moduleRoot -ChildPath "bin/$osName-$archName/$binaryName"

    if (Test-Path -Path $candidate) {
        return $candidate
    }

    $existing = Get-Command -Name kubememo -ErrorAction SilentlyContinue
    if ($existing) {
        return $existing.Source
    }

    $go = Get-Command -Name go -ErrorAction SilentlyContinue
    if (-not $go) {
        throw "kubememo binary not found at $candidate and Go is not installed to build it."
    }

    $null = New-Item -ItemType Directory -Path (Split-Path -Parent $candidate) -Force
    $cmd = "go build -o `"$candidate`" ./cmd/kubememo"
    $result = & $go.Source -C $repoRoot build -o $candidate ./cmd/kubememo 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build kubememo binary: $($result -join [Environment]::NewLine)"
    }

    return $candidate
}
