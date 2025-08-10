# Generate DNS artifacts using the sb29guard container (easy-mode)
# Usage examples:
#   .\gen-dns.ps1 hosts a-record 10.10.10.50
#   .\gen-dns.ps1 bind cname blocked.guard.local
#   .\gen-dns.ps1 domain-list
#   .\gen-dns.ps1 dnsmasq a-record 10.10.10.50
#   .\gen-dns.ps1 winps cname blocked.guard.local

param(
  [Parameter(Mandatory=$true)][ValidateSet('hosts','bind','unbound','rpz','dnsmasq','domain-list','winps')]
  [string]$Format,
  [Parameter(Mandatory=$false)][ValidateSet('a-record','cname')]
  [string]$Mode,
  [Parameter(Mandatory=$false)]
  [string]$Redirect,
  [string]$OutFile
)

$composeFile = Join-Path $PSScriptRoot 'docker-compose.yml'
$policyDir = Join-Path $PSScriptRoot 'policy'
$outDir = Join-Path $PSScriptRoot 'out'
if (-not (Test-Path $composeFile)) { Write-Error "Compose file not found: $composeFile"; exit 1 }
if (-not (Test-Path $policyDir)) { Write-Error "Policy dir not found: $policyDir"; exit 1 }
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

# Default output file path per format
if (-not $OutFile) {
  switch ($Format) {
    'hosts' { $OutFile = Join-Path $outDir 'hosts.txt' }
    'bind' { $OutFile = Join-Path $outDir 'zone.db' }
    'unbound' { $OutFile = Join-Path $outDir 'unbound.conf' }
    'rpz' { $OutFile = Join-Path $outDir 'policy.rpz' }
    'dnsmasq' { $OutFile = Join-Path $outDir 'dnsmasq.conf' }
    'domain-list' { $OutFile = Join-Path $outDir 'domains.txt' }
    'winps' { $OutFile = Join-Path $outDir 'windows-dns.ps1' }
  }
}

# Build base command
$cmd = @('compose','-f', $composeFile, 'run','--rm','sb29guard','generate-dns','--policy','/app/policy/domains.yaml','--format', $Format)

if ($Mode) { $cmd += @('--mode', $Mode) }
if ($Redirect) {
  if ($Mode -eq 'a-record') { $cmd += @('--redirect-ipv4', $Redirect) }
  elseif ($Mode -eq 'cname') { $cmd += @('--redirect-host', $Redirect) }
}
$cmd += @('--out', ('/out/' + [System.IO.Path]::GetFileName($OutFile)))

Write-Host "Running: docker $($cmd -join ' ')"
$proc = Start-Process -FilePath 'docker' -ArgumentList $cmd -NoNewWindow -PassThru -Wait
if ($proc.ExitCode -ne 0) { Write-Error "DNS generation failed"; exit $proc.ExitCode }

Write-Host "Wrote: $OutFile"
