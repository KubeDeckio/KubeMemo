function Resolve-KubeMemoRuntimeStore {
    [CmdletBinding()]
    param(
        [string]$Namespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    $safeNamespaceNames = @('kubememo-runtime', 'kubedeck-runtime', 'operations-runtime')
    $isInstalled = (Test-KubeMemoCrdInstalled -Store Runtime) -and (Test-KubeMemoNamespaceInstalled -Namespace $Namespace)
    $namespaceJson = Invoke-KubeMemoKubectl -Arguments @('get', 'namespace', $Namespace, '-o', 'json') -IgnoreErrors
    $metadata = if ($LASTEXITCODE -eq 0 -and $namespaceJson) { (($namespaceJson -join "`n") | ConvertFrom-Json).metadata } else { $null }
    $annotations = if ($metadata) { $metadata.annotations } else { $null }
    $labels = if ($metadata) { $metadata.labels } else { $null }
    $safeForGitOps = $safeNamespaceNames -contains $Namespace `
        -or $annotations.'runtime.notes.kubememo.io/non-reconciled' -eq 'true' `
        -or $labels.'runtime.notes.kubememo.io/non-reconciled' -eq 'true'

    [pscustomobject]@{
        Enabled = $isInstalled
        Namespace = $Namespace
        SafeForGitOps = $safeForGitOps
        NamespaceLabels = $labels
        NamespaceAnnotations = $annotations
    }
}
