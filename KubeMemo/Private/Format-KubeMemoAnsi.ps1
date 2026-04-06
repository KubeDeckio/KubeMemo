function Format-KubeMemoAnsi {
    [CmdletBinding()]
    param(
        [AllowNull()]
        [string]$Text,

        [string]$Code,

        [switch]$Disable
    )

    if ($Disable -or [string]::IsNullOrEmpty($Code) -or [string]::IsNullOrEmpty($Text)) {
        return $Text
    }

    $esc = [char]27
    return "$esc[$Code" + "m$Text$esc[0m"
}
