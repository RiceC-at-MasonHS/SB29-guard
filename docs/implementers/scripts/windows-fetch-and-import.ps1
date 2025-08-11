# SB29-guard – Windows nightly fetch and proxy import
# Purpose: After sb29-guard refreshes its Google Sheet (default ~23:59),
#          fetch the latest domain list and import into your proxy, then reload.
#
# How to schedule (Task Scheduler)
# 1) Open Task Scheduler > Create Task (not Basic Task)
# 2) General: Name "SB29 domain list import"; Run whether user is logged on or not; Run with highest privileges
# 3) Triggers: New… Daily at 00:10 (or a few minutes after your sb29-guard --refresh-at time)
# 4) Actions: New… Start a program
#    Program/script: powershell.exe
#    Add arguments (one line): -ExecutionPolicy Bypass -File "C:\\path\\to\\windows-fetch-and-import.ps1" -GuardBase "https://guard.school.internal" -OutFile "C:\\sb29\\blocked.txt" -OnlyWhenChanged
# 5) Conditions/Settings: adjust as desired; click OK; provide credentials
#
# Customize the PROXY section to match your environment.

param(
  [string]$GuardBase = "https://guard.school.internal",
  [string]$OutFile   = "C:\\sb29\\blocked.txt",
  [switch]$OnlyWhenChanged = $true
)

$ErrorActionPreference = "Stop"

function Get-PolicyVersion {
  try {
    $m = Invoke-WebRequest -UseBasicParsing -Uri "$GuardBase/metrics" | Select-Object -ExpandProperty Content | ConvertFrom-Json
    return $m.policy_version
  } catch {
    return ""
  }
}

$prevVerFile = [System.IO.Path]::ChangeExtension($OutFile, ".ver")
$prevVer = if (Test-Path $prevVerFile) { Get-Content $prevVerFile -ErrorAction SilentlyContinue } else { "" }
$curVer = Get-PolicyVersion
if ($OnlyWhenChanged -and $curVer -and $curVer -eq $prevVer) {
  Write-Host "No policy change (version $curVer). Skipping."
  exit 0
}

Write-Host "Fetching domain list from $GuardBase/domain-list"
Invoke-WebRequest -UseBasicParsing -Uri "$GuardBase/domain-list" -OutFile $OutFile

<#
PROXY IMPORT/RELOAD EXAMPLES (pick one and customize)

1) NGINX on a Linux host (copy list then reload via OpenSSH client on Windows 10+)
   $proxyHost = "proxy.example"  # SSH reachable host running NGINX
   $remotePath = "/etc/nginx/sb29/blocked.txt"
   scp $OutFile "$proxyHost:$remotePath"
   ssh $proxyHost 'nginx -s reload'

2) Apache httpd on Windows (if running locally)
   & "C:\Program Files\Apache24\bin\httpd.exe" -k restart

3) HAProxy Runtime API (TCP socket enabled in haproxy.cfg: e.g., "stats socket ipv4@127.0.0.1:9999 level admin")
   # This helper clears and repopulates a map from your file without a full reload.
   function Send-HAProxyCmd([string]$Host, [int]$Port, [string]$Cmd) {
     try {
       $client = [System.Net.Sockets.TcpClient]::new($Host,$Port)
       $stream = $client.GetStream()
       $writer = New-Object System.IO.StreamWriter($stream)
       $writer.NewLine = "`n"; $writer.AutoFlush = $true
       $writer.WriteLine($Cmd)
       $writer.Flush()
       $stream.Dispose(); $client.Close()
     } catch { Write-Warning "HAProxy cmd failed: $Cmd :: $_" }
   }
   function Update-HAProxyMapFromFile([string]$Host, [int]$Port, [string]$MapPath, [string]$FilePath) {
     Write-Host "Clearing HAProxy map $MapPath on $Host:$Port"
     Send-HAProxyCmd $Host $Port "clear map $MapPath"
     Get-Content $FilePath | ForEach-Object {
       $d = $_.Trim(); if (-not [string]::IsNullOrWhiteSpace($d)) {
         Send-HAProxyCmd $Host $Port ("add map {0} {1} 1" -f $MapPath,$d)
       }
     }
     Write-Host "HAProxy map update complete."
   }
   # Example usage:
   # Update-HAProxyMapFromFile -Host "127.0.0.1" -Port 9999 -MapPath "/etc/haproxy/blocked.map" -FilePath $OutFile

4) Vendor GUI proxies (import API)
   # Check your product's API or CLI for importing a domain list, then call here.
   # Example placeholder:
   # Invoke-RestMethod -Method POST -Uri https://proxy.local/api/blocklist/import -InFile $OutFile -ContentType 'text/plain'
#>

# Persist current version
if ($curVer) { Set-Content -Path $prevVerFile -Value $curVer }

Write-Host "Done."
