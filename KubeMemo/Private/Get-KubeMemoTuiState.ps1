function Get-KubeMemoTuiState {
    [CmdletBinding()]
    param(
        [string[]]$Namespaces,
        [switch]$IncludeRuntime,
        [string]$RuntimeNamespace,
        [string]$Kind,
        [string]$Text
    )

    $notes = @(Get-KubeMemo -IncludeRuntime:$IncludeRuntime -RuntimeNamespace $RuntimeNamespace -Namespace $Namespaces)
    if ($Kind -or $Text) {
        $notes = @($notes | Get-KubeMemoMatches -Kind $Kind -Text $Text)
    }
    $notes = @($notes | Sort-Object StoreType, Namespace, Kind, Name, Title)
    $visibleNamespaces = @($notes | Where-Object Namespace | Select-Object -ExpandProperty Namespace -Unique | Sort-Object)
    $visibleKinds = @($notes | Where-Object Kind | Select-Object -ExpandProperty Kind -Unique | Sort-Object)

    [pscustomobject]@{
        Notes = $notes
        VisibleNamespaces = $visibleNamespaces
        VisibleKinds = $visibleKinds
    }
}
