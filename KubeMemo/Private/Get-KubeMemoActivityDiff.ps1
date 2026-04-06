function Get-KubeMemoActivityDiff {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [hashtable]$OldObject,

        [Parameter(Mandatory)]
        [hashtable]$NewObject
    )

    $changes = New-Object System.Collections.Generic.List[object]

    $mappings = @(
        @{ Action = 'scale'; Path = 'spec.replicas'; Old = $OldObject.spec.replicas; New = $NewObject.spec.replicas }
        @{ Action = 'serviceTypeChange'; Path = 'spec.type'; Old = $OldObject.spec.type; New = $NewObject.spec.type }
        @{ Action = 'imageChange'; Path = 'spec.template.spec.containers'; Old = (($OldObject.spec.template.spec.containers | ForEach-Object { "$($_.name)=$($_.image)" }) -join ','); New = (($NewObject.spec.template.spec.containers | ForEach-Object { "$($_.name)=$($_.image)" }) -join ',') }
        @{ Action = 'resourceChange'; Path = 'spec.template.spec.containers.resources'; Old = ((($OldObject.spec.template.spec.containers | ForEach-Object { $_.resources }) | ConvertTo-Json -Depth 8) -join ''); New = ((($NewObject.spec.template.spec.containers | ForEach-Object { $_.resources }) | ConvertTo-Json -Depth 8) -join '') }
        @{ Action = 'nodeSelectorChange'; Path = 'spec.template.spec.nodeSelector'; Old = ($OldObject.spec.template.spec.nodeSelector | ConvertTo-Json -Depth 5 -Compress); New = ($NewObject.spec.template.spec.nodeSelector | ConvertTo-Json -Depth 5 -Compress) }
        @{ Action = 'tolerationChange'; Path = 'spec.template.spec.tolerations'; Old = ($OldObject.spec.template.spec.tolerations | ConvertTo-Json -Depth 5 -Compress); New = ($NewObject.spec.template.spec.tolerations | ConvertTo-Json -Depth 5 -Compress) }
    )

    foreach ($mapping in $mappings) {
        if ($mapping.Old -ne $mapping.New) {
            $changes.Add([pscustomobject]@{
                Action = $mapping.Action
                FieldPath = $mapping.Path
                OldValue = [string]$mapping.Old
                NewValue = [string]$mapping.New
            })
        }
    }

    if (($OldObject.kind -eq 'Ingress' -or $NewObject.kind -eq 'Ingress')) {
        $oldIngress = $OldObject.spec.rules | ConvertTo-Json -Depth 8 -Compress
        $newIngress = $NewObject.spec.rules | ConvertTo-Json -Depth 8 -Compress
        if ($oldIngress -ne $newIngress) {
            $changes.Add([pscustomobject]@{
                Action = 'ingressChange'
                FieldPath = 'spec.rules'
                OldValue = [string]$oldIngress
                NewValue = [string]$newIngress
            })
        }
    }

    return $changes
}
