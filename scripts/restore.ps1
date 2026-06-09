# restore.ps1
# Restores original game files from backup AND copies modified files from output/.
# Run backup-pcc.ps1 first if you haven't already.

param(
    [string]$GameDir = "C:\Program Files\EA Games\Mass Effect 2\BioGame\CookedPC",
    [string]$BackupDir = "$env:USERPROFILE\Desktop\MassEffect2Backups",
    [string]$OutputDir = "$PSScriptRoot",
    [switch]$RestoreOnly
)

$ErrorActionPreference = "Stop"

# ------------------------------------------------------------------
# Files managed by this script.
# Add new entries as you modify more files.
# ------------------------------------------------------------------
$files = @(
    @{Name = "BioD_CitHub_400LowerWing_LOC_INT.pcc"; Modded = $true},
    @{Name = "BIOGame_INT.tlk";                    Modded = $false}
)

foreach ($entry in $files) {
    $f = $entry.Name
    $gamePath = Join-Path $GameDir $f
    $backupPath = Join-Path $BackupDir $f
    $moddedName = [IO.Path]::GetFileNameWithoutExtension($f) + "_POC" + [IO.Path]::GetExtension($f)
    $moddedPath = Join-Path $OutputDir $moddedName

    if ($RestoreOnly -or -not $entry.Modeded) {
        # Restore original from backup
        if (-not (Test-Path $backupPath)) {
            Write-Error "Backup not found: $backupPath. Run backup-pcc.ps1 first."
            exit 1
        }
        Copy-Item $backupPath $gamePath -Force
        Write-Host "Restored original: $f"
    } else {
        # Copy modified version
        if (Test-Path $moddedPath) {
            Copy-Item $moddedPath $gamePath -Force
            Write-Host "Installed modded:  $f  <-  $moddedName"
        } else {
            Write-Warning "Modded file not found: $moddedPath (restoring original instead)"
            Copy-Item $backupPath $gamePath -Force
        }
    }
}

Write-Host "Done."
