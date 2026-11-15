# Final package builder for a development Windows PC with official Windows SDK tools.
# No antivirus changes, no silent installation, no privilege escalation.
param(
  [string]$IdentityName = 'SamerTallauze.JaguarDesktopPet',
  [string]$IdentityPublisher = 'CN=C83AFC34-5CE3-4E0A-8DCD-9733097BE6AC',
  [string]$PublisherDisplayName = 'Samer Tallauze'
)
$ErrorActionPreference = 'Stop'
$expectedName = 'SamerTallauze.JaguarDesktopPet'
$expectedPublisher = 'CN=C83AFC34-5CE3-4E0A-8DCD-9733097BE6AC'
if ($IdentityName -cne $expectedName -or $IdentityPublisher -cne $expectedPublisher -or $PublisherDisplayName -cne 'Samer Tallauze') {
    throw 'Package identity is not identical to the identity captured from Partner Center. Re-check the Product identity page.'
}
$root = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$payload = Join-Path $root 'payload-for-makeappx'
$outdir = Join-Path $root 'build'
$out = Join-Path $outdir 'SamerTallauze.JaguarDesktopPet_0.1.0.0_x64_OFFICIAL_MAKEAPPX_UNSIGNED.msix'
if (!(Test-Path (Join-Path $root 'app/JaguarPet.exe'))) { throw 'Missing native app executable.' }
if (Test-Path $payload) { Remove-Item $payload -Recurse -Force }
New-Item -ItemType Directory -Force -Path (Join-Path $payload 'Assets') | Out-Null
New-Item -ItemType Directory -Force -Path $outdir | Out-Null
Copy-Item (Join-Path $root 'app/JaguarPet.exe') (Join-Path $payload 'JaguarPet.exe') -Force
Copy-Item (Join-Path $root 'store-assets/*.png') (Join-Path $payload 'Assets') -Force
$manifest = Get-Content (Join-Path $root 'packaging/AppxManifest.template.xml') -Raw -Encoding UTF8
$manifest = $manifest.Replace('__IDENTITY_NAME__', [System.Security.SecurityElement]::Escape($IdentityName))
$manifest = $manifest.Replace('__IDENTITY_PUBLISHER__', [System.Security.SecurityElement]::Escape($IdentityPublisher))
$manifest = $manifest.Replace('__PUBLISHER_DISPLAY_NAME__', [System.Security.SecurityElement]::Escape($PublisherDisplayName))
[xml]$check = $manifest
[System.IO.File]::WriteAllText((Join-Path $payload 'AppxManifest.xml'), $manifest, (New-Object System.Text.UTF8Encoding($false)))
$makeAppx = (Get-Command 'makeappx.exe' -ErrorAction SilentlyContinue | Select-Object -First 1 -ExpandProperty Source)
if (-not $makeAppx) {
    $sdkDir = Join-Path ${env:ProgramFiles(x86)} 'Windows Kits\10\bin'
    if (Test-Path $sdkDir) {
        $makeAppx = (Get-ChildItem $sdkDir -Recurse -Filter MakeAppx.exe -ErrorAction SilentlyContinue |
            Where-Object { $_.FullName -match '[\\/]x64[\\/]' } |
            Sort-Object FullName -Descending | Select-Object -First 1 -ExpandProperty FullName)
    }
}
if (-not $makeAppx) { throw 'MakeAppx.exe not found. Install the Windows SDK Packaging Tools on an authorized Windows development machine.' }
if (Test-Path $out) { Remove-Item $out -Force }
Write-Host "Building official unsigned MSIX with $makeAppx"
& $makeAppx pack /d $payload /p $out /o /l
if ($LASTEXITCODE -ne 0) { throw "MakeAppx pack returned $LASTEXITCODE" }
if (Test-Path (Join-Path $root 'makeappx-validation-unpacked')) { Remove-Item (Join-Path $root 'makeappx-validation-unpacked') -Recurse -Force }
& $makeAppx unpack /p $out /d (Join-Path $root 'makeappx-validation-unpacked') /o
if ($LASTEXITCODE -ne 0) { throw "MakeAppx unpack returned $LASTEXITCODE" }
Get-FileHash $out -Algorithm SHA256 | Format-List
Write-Host "DONE: $out"
Write-Warning 'Unsigned: not an installer for a managed company PC. Before Store submission, complete Windows runtime tests and WACK.'
