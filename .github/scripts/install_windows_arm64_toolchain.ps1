param([switch]$FixedRelease)
$ErrorActionPreference = "Stop"
$fallback = @{
    name = "llvm-mingw-20261006-ucrt-x86_64.zip"
    browser_download_url = "https://github.com/mstorsjo/llvm-mingw/releases/download/20261006/llvm-mingw-20261006-ucrt-x86_64.zip"
    size = 190836485
    digest = "sha256:317492c456aa27ee607a5919f1d2d38dcdc1112516a24d0bf4b00d078f52d17a"
}
$candidates = @()
try {
    if ($FixedRelease) { throw "Fixed release selected for recovery verification" }
    $headers = @{ Accept = "application/vnd.github+json" }
    if ($env:GH_TOKEN) { $headers.Authorization = "Bearer $env:GH_TOKEN" }
    $release = Invoke-RestMethod -TimeoutSec 60 -MaximumRetryCount 2 -RetryIntervalSec 2 `
      -Headers $headers -Uri "https://api.github.com/repos/mstorsjo/llvm-mingw/releases/latest"
    $latest = $release.assets | Where-Object {
        $_.name -match '^llvm-mingw-[0-9]+-ucrt-x86_64\.zip$' -and
        $_.digest -match '^sha256:[0-9a-f]{64}$' -and
        $_.browser_download_url -match '^https://github\.com/mstorsjo/llvm-mingw/releases/download/'
    } | Select-Object -First 1
    if ($latest) { $candidates += $latest }
} catch {
    Write-Host "Toolchain metadata unavailable; using the verified fixed release."
}
$candidates += $fallback
$compiler = $null
foreach ($asset in $candidates) {
    $archive = Join-Path $env:RUNNER_TEMP $asset.name
    $destination = Join-Path $env:RUNNER_TEMP "llvm-mingw"
    try {
        if (Test-Path $destination) { Remove-Item -Recurse -Force $destination }
        Invoke-WebRequest -TimeoutSec 1200 -MaximumRetryCount 2 -RetryIntervalSec 2 `
          -Uri $asset.browser_download_url -OutFile $archive
        if ((Get-Item $archive).Length -ne $asset.size) { throw "Toolchain archive size mismatch" }
        $digest = (Get-FileHash -Algorithm SHA256 $archive).Hash.ToLowerInvariant()
        if ("sha256:$digest" -ne $asset.digest) { throw "Toolchain archive digest mismatch" }
        Expand-Archive -Path $archive -DestinationPath $destination
        $compiler = Get-ChildItem -Path $destination -Recurse -Filter "aarch64-w64-mingw32-clang.exe" | Select-Object -First 1
        if (-not $compiler) { throw "ARM64 compiler missing in toolchain archive" }
        & $compiler.FullName --version
        if ($LASTEXITCODE -ne 0) { throw "ARM64 compiler self-check failed" }
        break
    } catch {
        $compiler = $null
        Write-Host "Toolchain candidate failed verification; trying the fixed release."
    } finally {
        Remove-Item -Force $archive -ErrorAction SilentlyContinue
    }
}
if (-not $compiler) { throw "No verified Windows ARM64 toolchain is available" }
$compiler.Directory.FullName | Out-File -FilePath $env:GITHUB_PATH -Encoding utf8 -Append
"CC=aarch64-w64-mingw32-clang" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append
"CXX=aarch64-w64-mingw32-clang++" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append
"CGO_ENABLED=1" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append
