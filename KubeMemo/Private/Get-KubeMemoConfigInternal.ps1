function Get-KubeMemoConfigInternal {
    [CmdletBinding()]
    param()

    [pscustomobject]@{
        Enabled         = $true
        RuntimeNamespace = 'kubememo-runtime'
        RuntimeDefaultExpiry = '24h'
        IncidentDefaultExpiry = '12h'
        GitOpsRepoPath  = './ops/kubememo'
        DedupeWindow    = '60s'
        DurableCrdName  = 'memos.notes.kubememo.io'
        RuntimeCrdName  = 'runtimememos.runtime.notes.kubememo.io'
        Rendering = [pscustomobject]@{
            Cards = $true
            Markdown = $true
            Ansi = $true
        }
    }
}
