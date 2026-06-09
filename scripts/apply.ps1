# apply.ps1
# Copies modified files from output/ into the game folder.
# Run backup-pcc.ps1 first if you haven't already.

param(
    [string]$GameDir = "C:\Program Files\EA Games\Mass Effect 2\BioGame\CookedPC",
    [string]$OutputDir = "$PSScriptRoot\..\output"
)

$ErrorActionPreference = "Stop"

# ------------------------------------------------------------------
#   Name     = original filename in the game folder
#   ModName  = modified filename in output/
# ------------------------------------------------------------------
$files = @(
    @{Name = "BioD_CitHub_400LowerWing_LOC_INT.pcc"; ModName = "BioD_CitHub_400LowerWing_POC.pcc"}
)

foreach ($entry in $files) {
    $gamePath = Join-Path $GameDir $entry.Name
    $moddedPath = Join-Path $OutputDir $entry.ModName

    if (-not (Test-Path $moddedPath)) {
        Write-Error "Modded file not found: $moddedPath"
        exit 1
    }
    Copy-Item $moddedPath $gamePath -Force
    Write-Host "Applied: $($entry.Name)  <-  $($entry.ModName)"
}

Write-Host "Done."
