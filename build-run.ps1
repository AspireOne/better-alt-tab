$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$outputDir = Join-Path $repoRoot ".gotmp"
$outputExe = Join-Path $outputDir "quick-app-switcher.exe"

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

Push-Location $repoRoot
try {
    go build -o $outputExe ./cmd/quick-app-switcher
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    & $outputExe
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
