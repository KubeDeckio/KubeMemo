function Test-KubeMemoInstallation {
    [CmdletBinding()]
    param(
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    $null = Invoke-KubeMemoKubectl -Arguments @('cluster-info') -IgnoreErrors
    [pscustomobject]@{
        ClusterReachable = ($LASTEXITCODE -eq 0)
        DurableCrdInstalled = Test-KubeMemoCrdInstalled -Store Durable
        RuntimeCrdInstalled = Test-KubeMemoCrdInstalled -Store Runtime
        RuntimeNamespaceInstalled = Test-KubeMemoNamespaceInstalled -Namespace $RuntimeNamespace
        RbacInstalled = Test-KubeMemoRbacInstalled
        GitOps = Resolve-KubeMemoGitOpsState
        RuntimeStore = Resolve-KubeMemoRuntimeStore -Namespace $RuntimeNamespace
    }
}
