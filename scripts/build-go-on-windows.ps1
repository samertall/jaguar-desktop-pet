$ErrorActionPreference='Stop'
$root=(Resolve-Path (Join-Path $PSScriptRoot '..')).Path
Push-Location (Join-Path $root 'app')
try {
  $env:GOOS='windows'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
  go test motion.go motion_test.go
  if ($LASTEXITCODE -ne 0) { throw 'Go tests failed' }
  go build -trimpath -ldflags '-H windowsgui -s -w' -o JaguarPet.exe .
  if ($LASTEXITCODE -ne 0) { throw 'Go build failed' }
} finally { Pop-Location }
