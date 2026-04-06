$publicPath = Join-Path -Path $PSScriptRoot -ChildPath 'Public'
$privatePath = Join-Path -Path $PSScriptRoot -ChildPath 'Private'

foreach ($script in Get-ChildItem -Path $privatePath -Filter '*.ps1' -File | Sort-Object Name) {
    . $script.FullName
}

foreach ($script in Get-ChildItem -Path $publicPath -Filter '*.ps1' -File | Sort-Object Name) {
    . $script.FullName
}
