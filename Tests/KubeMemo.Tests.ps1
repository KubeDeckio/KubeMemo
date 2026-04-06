Import-Module "$PSScriptRoot/../KubeMemo/KubeMemo.psd1" -Force

Describe 'KubeMemo module' {
    It 'exports the expected public commands' {
        $commands = Get-Command -Module KubeMemo | Select-Object -ExpandProperty Name
        $commands | Should -Contain 'Install-KubeMemo'
        $commands | Should -Contain 'New-KubeMemo'
        $commands | Should -Contain 'Show-KubeMemo'
        $commands | Should -Contain 'Open-KubeMemoTui'
        $commands | Should -Contain 'Sync-KubeMemoGitOps'
    }

    It 'returns default configuration' {
        $config = Get-KubeMemoConfig
        $config.RuntimeNamespace | Should -Be 'kubememo-runtime'
        $config.RuntimeDefaultExpiry | Should -Be '24h'
    }
}

Describe 'Private helpers' {
    InModuleScope KubeMemo {
        It 'resolves a resource target' {
            $target = Resolve-KubeMemoTarget -Kind Deployment -Name orders-api -Namespace prod -ApiVersion apps/v1
            $target.mode | Should -Be 'resource'
            $target.kind | Should -Be 'Deployment'
            $target.namespace | Should -Be 'prod'
            $target.name | Should -Be 'orders-api'
        }

        It 'creates activity diffs for replica changes' {
            $oldObject = @{
                spec = @{
                    replicas = 2
                }
            }
            $newObject = @{
                spec = @{
                    replicas = 5
                }
            }

            $diff = Get-KubeMemoActivityDiff -OldObject $oldObject -NewObject $newObject
            $diff.Action | Should -Contain 'scale'
        }

        It 'renders a simple note card' {
            $card = Write-KubeMemoCard -Notes @(
                [pscustomobject]@{
                    Kind = 'Deployment'
                    Namespace = 'prod'
                    Name = 'orders-api'
                    StoreType = 'Durable'
                    Title = 'Orders API warm-up behavior'
                    NoteType = 'advisory'
                    OwnerTeam = 'platform'
                    OwnerContact = '@platform'
                    Summary = 'Expected transient 502s'
                    Content = 'Ignore for 3 minutes.'
                    SourceType = 'manual'
                    SourceGenerator = ''
                    Confidence = ''
                    Activity = $null
                    ExpiresAt = $null
                }
            ) -PassThru

            $card | Should -Match 'KubeMemo: Deployment/prod/orders-api'
            $card | Should -Match 'Durable Notes'
        }

        It 'serializes a resource to yaml' {
            $yaml = ConvertTo-KubeMemoYaml -Resource @{
                apiVersion = 'notes.kubememo.io/v1alpha1'
                kind = 'Memo'
                metadata = @{
                    name = 'orders-api'
                    namespace = 'prod'
                }
                spec = @{
                    title = 'Orders API'
                    content = "line1`nline2"
                }
            }

            $yaml | Should -Match 'apiVersion: notes.kubememo.io/v1alpha1'
            $yaml | Should -Match 'kind: Memo'
            $yaml | Should -Match 'content:'
        }
    }
}
