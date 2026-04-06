function Test-KubeMemoDurableWriteAllowed {
    [CmdletBinding()]
    param()

    $gitOps = Resolve-KubeMemoGitOpsState
    return (-not $gitOps.Enabled)
}
