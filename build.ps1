param([switch]$Clean)

$ErrorActionPreference = "Stop"

$APP_NAME = "BPB-Wizard"
$OUT_DIR  = "bin"
$DIST_DIR = "dist"
$GOOS     = "windows"
$GOARCH   = "amd64"

if ($Clean) {
    if (Test-Path $OUT_DIR)  { Remove-Item -Recurse -Force $OUT_DIR }
    if (Test-Path $DIST_DIR) { Remove-Item -Recurse -Force $DIST_DIR }
    Write-Host "Cleaned $OUT_DIR and $DIST_DIR"
    exit 0
}

$VERSION    = (Get-Content -Path "$PSScriptRoot\VERSION" -First 1).Trim()
$timestamp  = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd HH:mm:ss")
$goVerRaw   = & go version
$goVersion  = if ($goVerRaw -match 'go(\d+\.\d+(?:\.\d+)?)') { $Matches[1] } else { "unknown" }

$LDFLAGS = "-s -w -X `"main.BuildTimestamp=$timestamp`" -X `"main.VERSION=$VERSION`" -X `"main.goVersion=$goVersion`""

$outDir = "$OUT_DIR\$APP_NAME-$GOOS-$GOARCH"
if (-not (Test-Path $outDir))  { New-Item -ItemType Directory -Path $outDir -Force | Out-Null }
if (-not (Test-Path $DIST_DIR)) { New-Item -ItemType Directory -Path $DIST_DIR -Force | Out-Null }

Write-Host "Building for $GOOS/$GOARCH..."

$env:GO111MODULE = "on"
$env:CGO_ENABLED = "0"
$env:GOOS        = $GOOS
$env:GOARCH      = $GOARCH

& go build -trimpath -ldflags $LDFLAGS -o "$outDir\Wizard.exe"

if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed"
    exit 1
}

Copy-Item "$PSScriptRoot\LICENSE" $outDir

$archiveFile = "$DIST_DIR\$APP_NAME-$GOOS-$GOARCH.zip"
if (Test-Path $archiveFile) { Remove-Item $archiveFile }
Compress-Archive -Path "$outDir\*" -DestinationPath $archiveFile

Write-Host "Done -> $archiveFile"
