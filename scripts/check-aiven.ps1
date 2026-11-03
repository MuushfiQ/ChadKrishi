$ErrorActionPreference = 'Stop'

$hostName = Read-Host 'Aiven host [mysql-364be7d0-chadkrishi.d.aivencloud.com]'
if ([string]::IsNullOrWhiteSpace($hostName)) {
  $hostName = 'mysql-364be7d0-chadkrishi.d.aivencloud.com'
}

$userName = Read-Host 'Aiven username [avnadmin]'
if ([string]::IsNullOrWhiteSpace($userName)) {
  $userName = 'avnadmin'
}

$caFile = Read-Host 'Path to Aiven CA PEM file'
if (-not (Test-Path -LiteralPath $caFile -PathType Leaf)) {
  throw "CA certificate was not found: $caFile"
}

$password = Read-Host 'Aiven password (input is hidden)' -AsSecureString
$passwordText = [System.Net.NetworkCredential]::new('', $password).Password

$oldEnvironment = @{
  APP_ENV = $env:APP_ENV
  DATABASE_DSN = $env:DATABASE_DSN
  DATABASE_TLS_SERVER_NAME = $env:DATABASE_TLS_SERVER_NAME
  DATABASE_TLS_CA_FILE = $env:DATABASE_TLS_CA_FILE
}

try {
  $env:APP_ENV = 'development'
  $env:DATABASE_DSN = "$userName`:$passwordText@tcp($hostName`:23268)/defaultdb?parseTime=true&charset=utf8mb4&tls=aiven"
  $env:DATABASE_TLS_SERVER_NAME = $hostName
  $env:DATABASE_TLS_CA_FILE = (Resolve-Path -LiteralPath $caFile).Path

  Set-Location (Join-Path $PSScriptRoot '..\backend')
  go run . -check-db
}
finally {
  foreach ($name in $oldEnvironment.Keys) {
    Set-Item -Path "Env:$name" -Value $oldEnvironment[$name]
  }
  $passwordText = $null
  $password = $null
}
