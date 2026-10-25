<#
.SYNOPSIS
    Deterministic performance measurement (pkg/neat/perf_test.go), used before and after every fix from raport_wydajnosc.md.

.EXAMPLE
    .\scripts\perf.ps1 -Label before            # measure and save to .ai_reports/perf/before.*
    .\scripts\perf.ps1 -Label after
    .\scripts\perf.ps1 -Compare before,after    # compare two saved measurements
#>
param(
    [string]$Label,
    [string[]]$Compare,
    [int]$Count = 10
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$outDir = Join-Path $root '.ai_reports/perf'
New-Item -ItemType Directory -Force $outDir | Out-Null

if ($Compare) {
    if ($Compare.Count -ne 2) { throw 'Compare needs exactly two labels, e.g. -Compare before,after' }
    $old, $new = $Compare
    foreach ($l in $old, $new) {
        Write-Host "===== ${l}: work (exact) =====" -ForegroundColor Cyan
        Get-Content (Join-Path $outDir "$l.work.txt")
    }
    Write-Host "===== time and allocations (benchstat) =====" -ForegroundColor Cyan
    Push-Location $outDir
    try { benchstat "$old=$old.bench.txt" "$new=$new.bench.txt" } finally { Pop-Location }
    exit
}

if (-not $Label) {
    $Label = (git -C $root rev-parse --short HEAD).Trim()
    if (git -C $root status --porcelain) { $Label += '-dirty' }
}

Push-Location $root
try {
    Write-Host "Counting activations ($Label)..." -ForegroundColor Cyan
    $work = go test ./pkg/neat -run '^TestPerfWorkload$' -count=1 -v
    if ($LASTEXITCODE -ne 0) { $work; throw 'TestPerfWorkload failed' }
    # keep only the report printed by t.Log
    $report = $work | Where-Object { $_ -match '^\s{8}' } | ForEach-Object { $_.Substring(8) }
    $header = "commit: $((git rev-parse --short HEAD).Trim())  date: $(Get-Date -Format 'yyyy-MM-dd HH:mm')"
    @($header) + $report | Set-Content -Encoding utf8 (Join-Path $outDir "$Label.work.txt")
    $report

    Write-Host "Benchmark, $Count runs ($Label)..." -ForegroundColor Cyan
    go test ./pkg/neat -run '^$' -bench '^BenchmarkPerfWorkload$' -benchmem -count $Count |
        Tee-Object -FilePath (Join-Path $outDir "$Label.bench.txt")
    if ($LASTEXITCODE -ne 0) { throw 'BenchmarkPerfWorkload failed' }
} finally {
    Pop-Location
}
Write-Host "Saved to $outDir\$Label.work.txt and $Label.bench.txt" -ForegroundColor Green
