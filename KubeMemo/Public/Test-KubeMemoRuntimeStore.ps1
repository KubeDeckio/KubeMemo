function Test-KubeMemoRuntimeStore {
    [CmdletBinding()]
    param(
        [string]$Namespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    Resolve-KubeMemoRuntimeStore -Namespace $Namespace
}
