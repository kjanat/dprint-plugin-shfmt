<#
For my personal machine only... for now I guess...
#>

[CmdletBinding()]
param(
	[string]$TinyGoPath = "$env:LOCALAPPDATA\Microsoft\WinGet\Packages\tinygo-org.tinygo_Microsoft.Winget.Source_8wekyb3d8bbwe\tinygo\bin\tinygo.exe",
	[string]$WasmOptPath = "$env:USERPROFILE\scoop\shims\wasm-opt.exe",
	[string]$Version = "0.0.5",
	[string]$ReleaseTag = "0.0.5",
	[string]$RepoSlug = "kjanat/shfmt",
	[string]$GitHubRepo = "kjanat/dprint-plugin-shfmt",
	[string]$OutputPath = "plugin.wasm",
	[uint32]$StackSize = 1MB
)

$ErrorActionPreference = "Stop"

function Resolve-ToolPath
{
	param(
		[string]$ToolPath,
		[string]$ToolName
	)

	if (Test-Path $ToolPath)
	{
		return (Resolve-Path $ToolPath).Path
	}

	$command = Get-Command $ToolPath -ErrorAction SilentlyContinue
	if ($null -ne $command)
	{
		return $command.Source
	}

	throw "$ToolName not found: $ToolPath"
}

$tinyGoExe = Resolve-ToolPath -ToolPath $TinyGoPath -ToolName "TinyGo"
$wasmOptExe = Resolve-ToolPath -ToolPath $WasmOptPath -ToolName "wasm-opt"

$env:TINYGOROOT = Split-Path (Split-Path $tinyGoExe)
$env:WASMOPT = $wasmOptExe

Write-Host "TinyGo:     $tinyGoExe"
Write-Host "TINYGOROOT: $env:TINYGOROOT"
Write-Host "WASMOPT:    $env:WASMOPT"
Write-Host "Version:    $Version"
Write-Host "ReleaseTag: $ReleaseTag"
Write-Host "RepoSlug:   $RepoSlug"
Write-Host "GitHubRepo: $GitHubRepo"
Write-Host "Output:     $OutputPath"

$ldflags = @(
	"-extldflags '-z stack-size=$StackSize'"
	"-X main.Version=$Version"
	"-X main.ReleaseTag=$ReleaseTag"
	"-X main.RepoSlug=$RepoSlug"
	"-X main.GitHubRepo=$GitHubRepo"
) -join ' '

$buildArgs = @(
	"build"
	"-o", $OutputPath
	"-target=wasm-unknown"
	"-gc=conservative"
	"-scheduler=none"
	"-panic=trap"
	"-no-debug"
	"-ldflags=$ldflags"
	"."
)

& $tinyGoExe @buildArgs
if ($LASTEXITCODE -ne 0)
{
	throw "TinyGo build failed with exit code $LASTEXITCODE"
}

Write-Host "Built $OutputPath"
