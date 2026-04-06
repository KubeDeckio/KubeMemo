function Resolve-KubeMemoTarget {
    [CmdletBinding(DefaultParameterSetName = 'Resource')]
    param(
        [Parameter(ParameterSetName = 'Resource', Mandatory)]
        [string]$Kind,

        [Parameter(ParameterSetName = 'Resource', Mandatory)]
        [string]$Name,

        [Parameter(ParameterSetName = 'Resource')]
        [string]$ApiVersion = 'v1',

        [Parameter(ParameterSetName = 'Resource')]
        [string]$Namespace,

        [Parameter(ParameterSetName = 'Namespace', Mandatory)]
        [string]$TargetNamespace,

        [Parameter(ParameterSetName = 'App', Mandatory)]
        [string]$AppName,

        [Parameter(ParameterSetName = 'App')]
        [string]$AppInstance
    )

    switch ($PSCmdlet.ParameterSetName) {
        'Namespace' {
            return @{
                mode = 'namespace'
                namespace = $TargetNamespace
                kind = 'Namespace'
                name = $TargetNamespace
                apiVersion = 'v1'
            }
        }
        'App' {
            return @{
                mode = 'app'
                appRef = @{
                    name = $AppName
                    instance = $AppInstance
                }
            }
        }
        default {
            return @{
                mode = 'resource'
                apiVersion = $ApiVersion
                kind = $Kind
                namespace = $Namespace
                name = $Name
            }
        }
    }
}
