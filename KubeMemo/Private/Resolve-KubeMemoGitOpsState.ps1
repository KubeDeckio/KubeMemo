function Resolve-KubeMemoGitOpsState {
    [CmdletBinding()]
    param()

    $providers = @()
    foreach ($namespace in 'flux-system', 'argocd') {
        $null = Invoke-KubeMemoKubectl -Arguments @('get', 'namespace', $namespace, '-o', 'name') -IgnoreErrors
        if ($LASTEXITCODE -eq 0) {
            $providers += if ($namespace -eq 'flux-system') { 'Flux' } else { 'ArgoCD' }
        }
    }

    [pscustomobject]@{
        Enabled  = $providers.Count -gt 0
        Provider = if ($providers.Count -gt 0) { $providers -join ',' } else { 'None' }
        DurableSourceOfTruth = if ($providers.Count -gt 0) { 'Git' } else { 'Cluster' }
    }
}
