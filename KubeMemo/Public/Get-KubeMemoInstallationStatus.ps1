function Get-KubeMemoInstallationStatus {
    [CmdletBinding()]
    param(
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    $status = Test-KubeMemoInstallation -RuntimeNamespace $RuntimeNamespace
    $mode = 'standard'
    if ($status.GitOps.Enabled -and -not $status.RuntimeStore.Enabled) {
        $mode = 'GitOps durable-only'
    }
    elseif ($status.GitOps.Enabled -and $status.RuntimeStore.Enabled) {
        $mode = 'GitOps with runtime store'
    }

    [pscustomobject]@{
        Mode = $mode
        ClusterReachable = $status.ClusterReachable
        DurableCrdInstalled = $status.DurableCrdInstalled
        RuntimeCrdInstalled = $status.RuntimeCrdInstalled
        RuntimeNamespaceInstalled = $status.RuntimeNamespaceInstalled
        RbacInstalled = $status.RbacInstalled
        GitOpsProvider = $status.GitOps.Provider
        RuntimeNamespace = $status.RuntimeStore.Namespace
        RuntimeStoreSafe = $status.RuntimeStore.SafeForGitOps
    }
}
