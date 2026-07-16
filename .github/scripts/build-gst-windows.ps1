Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$ServerRoot = Join-Path $RepoRoot "server"
$GstLib = Join-Path $ServerRoot "gstreamer\gst-libs\win-x86_64"
$EmbedZip = Join-Path $ServerRoot "gstreamer\embedded_gstlib_windows_amd64.zip"
$Output = Join-Path $RepoRoot "dist\TorrServer-gst-windows-amd64.exe"

$RequiredRuntimePaths = @(
    "bin\gst-discoverer-1.0.exe",
    "bin\libgstreamer-1.0-0.dll",
    "bin\libgstapp-1.0-0.dll",
    "lib\gstreamer-1.0",
    "libexec\gstreamer-1.0\gst-plugin-scanner.exe"
)
foreach ($relativePath in $RequiredRuntimePaths) {
    $path = Join-Path $GstLib $relativePath
    if (!(Test-Path -LiteralPath $path)) {
        throw "Incomplete bundled GStreamer runtime: $path"
    }
}

$Ldflags = "-s -w -checklinkname=0"
if (![string]::IsNullOrWhiteSpace($env:VERSION)) {
    $Ldflags += " -X server/version.Version=$env:VERSION"
}

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Output) | Out-Null
Add-Type -AssemblyName System.IO.Compression.FileSystem

$OldGOOS = $env:GOOS
$OldGOARCH = $env:GOARCH
$OldCGOEnabled = $env:CGO_ENABLED

try {
    if (Test-Path -LiteralPath $EmbedZip) {
        Remove-Item -LiteralPath $EmbedZip -Force
    }

    [System.IO.Compression.ZipFile]::CreateFromDirectory(
        $GstLib,
        $EmbedZip,
        [System.IO.Compression.CompressionLevel]::Optimal,
        $false
    )

    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"

    Push-Location $ServerRoot
    try {
        & go build `
            -tags "nosqlite gst embed_gstlib" `
            -trimpath `
            -ldflags $Ldflags `
            -o $Output `
            .\cmd
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed with exit code $LASTEXITCODE"
        }
    } finally {
        Pop-Location
    }
} finally {
    if (Test-Path -LiteralPath $EmbedZip) {
        Remove-Item -LiteralPath $EmbedZip -Force
    }
    $env:GOOS = $OldGOOS
    $env:GOARCH = $OldGOARCH
    $env:CGO_ENABLED = $OldCGOEnabled
}

$Result = Get-Item -LiteralPath $Output
Write-Host "Built $($Result.FullName) ($([math]::Round($Result.Length / 1MB, 2)) MB)"
