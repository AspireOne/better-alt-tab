$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$outputDir = Join-Path $repoRoot ".gotmp"
$outputExe = Join-Path $outputDir "better-alt-tab.exe"

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

Push-Location $repoRoot
try {
    go build -ldflags "-H=windowsgui" -o $outputExe ./cmd/better-alt-tab
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    & $outputExe
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
