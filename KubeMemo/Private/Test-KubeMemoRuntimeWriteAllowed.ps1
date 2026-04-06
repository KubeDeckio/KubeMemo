function Test-KubeMemoRuntimeWriteAllowed {
    [CmdletBinding()]
    param(
        [string]$Namespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    $gitOps = Resolve-KubeMemoGitOpsState
    $runtime = Resolve-KubeMemoRuntimeStore -Namespace $Namespace

    if (-not $runtime.Enabled) {
        return $false
    }

    if ($gitOps.Enabled -and -not $runtime.SafeForGitOps) {
        return $false
    }

    return $true
}
