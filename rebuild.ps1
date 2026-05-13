<#
.SYNOPSIS
	Build dprint-plugin-shfmt.wasm with fork-identity ldflag injection.

.DESCRIPTION
	Local Windows rebuild for the fork's wasm artifact. Pinned to my
	personal toolchain layout (winget tinygo, scoop wasm-opt), but every
	path is overridable via parameter.

	Version, ReleaseTag, RepoSlug, GitHubRepo are validated against the
	fork convention: bare semver (no v-prefix), `owner/name` slug shape.

	Version defaults from `git describe --tags --abbrev=0`; GitHubRepo
	defaults from `git remote get-url origin`. Outside a git repo, both
	fall back to sane fork defaults.

.EXAMPLE
	.\rebuild.ps1
	Build plugin.wasm with the current default version.

.EXAMPLE
	.\rebuild.ps1 -Verbose
	Build the version derived from `git describe`, with tool diagnostics.

.EXAMPLE
	.\rebuild.ps1 -Version 0.0.8
	Override the git-derived version (ReleaseTag follows automatically).

.EXAMPLE
	.\rebuild.ps1 -Force
	Overwrite an existing plugin.wasm without prompting.

.EXAMPLE
	.\rebuild.ps1 -WhatIf
	Show what would happen without touching disk.
#>

#Requires -Version 7

[CmdletBinding(SupportsShouldProcess)]
[OutputType([System.IO.FileInfo])]
param(
	[ValidateNotNullOrEmpty()]
	[string]$TinyGoPath = (Join-Path $env:LOCALAPPDATA "Microsoft\WinGet\Packages\tinygo-org.tinygo_Microsoft.Winget.Source_8wekyb3d8bbwe\tinygo\bin\tinygo.exe"),

	[ValidateNotNullOrEmpty()]
	[string]$WasmOptPath = (Join-Path $env:USERPROFILE "scoop\shims\wasm-opt.exe"),

	[ValidatePattern('^\d+\.\d+\.\d+(-[\w.]+)?$')]
	[string]$Version = $(
		try {
			$tag = git describe --tags --abbrev=0 2>$null;
			if ($LASTEXITCODE -eq 0 -and $tag) {
				$tag.Trim('v', ' ')
			} else {
				'0.0.0-dev'
			}
		} catch {
			'0.0.0-dev'
		}
	),

	[ValidatePattern('^\d+\.\d+\.\d+(-[\w.]+)?$')]
	[string]$ReleaseTag = $Version,

	[ValidatePattern('^[\w.-]+/[\w.-]+$')]
	[string]$GitHubRepo = $(
		try {
			$url = git remote get-url origin 2>$null
			if ($LASTEXITCODE -eq 0 -and $url -and $url -match 'github\.com[:/](?<repo>[\w.-]+/[\w.-]+?)(?:\.git)?/?\s*$') {
				$matches.repo
			} else {
				'kjanat/dprint-plugin-shfmt'
			}
		} catch {
			'kjanat/dprint-plugin-shfmt'
		}
	),

	[ValidatePattern('^[\w.-]+/[\w.-]+$')]
	[string]$RepoSlug = $(
		if ($GitHubRepo -match '^(?<owner>[\w.-]+)/dprint-plugin-(?<name>.+)$') {
			"$($matches.owner)/$($matches.name)"
		} else {
			$GitHubRepo
		}
	),

	[ValidateNotNullOrEmpty()]
	[string]$OutputPath = "plugin.wasm",

	[ValidateRange(256KB, 16MB)]
	[uint32]$StackSize = 1MB,

	[switch]$Force
)

$ErrorActionPreference = "Stop"

function Resolve-ToolPath {
	param(
		[string]$ToolPath,
		[string]$ToolName
	)

	if (Test-Path $ToolPath) {
		return (Resolve-Path $ToolPath).Path
	}

	$command = Get-Command $ToolPath -ErrorAction SilentlyContinue
	if ($null -ne $command) {
		return $command.Source
	}

	throw "$ToolName not found: $ToolPath"
}

$tinyGoExe = Resolve-ToolPath -ToolPath $TinyGoPath -ToolName "TinyGo"
$wasmOptExe = Resolve-ToolPath -ToolPath $WasmOptPath -ToolName "wasm-opt"

$env:TINYGOROOT = Split-Path (Split-Path $tinyGoExe)
$env:WASMOPT = $wasmOptExe

Write-Verbose "TinyGo:     $tinyGoExe"
Write-Verbose "TINYGOROOT: $env:TINYGOROOT"
Write-Verbose "WASMOPT:    $env:WASMOPT"
Write-Verbose "Version:    $Version"
Write-Verbose "ReleaseTag: $ReleaseTag"
Write-Verbose "RepoSlug:   $RepoSlug"
Write-Verbose "GitHubRepo: $GitHubRepo"
Write-Verbose "Output:     $OutputPath"
Write-Verbose "StackSize:  $StackSize"

if ((Test-Path $OutputPath) -and -not ($Force -or $WhatIfPreference)) {
	if (-not $PSCmdlet.ShouldContinue("Overwrite existing $OutputPath?", "Build wasm")) {
		Write-Verbose "Aborted: $OutputPath already exists (pass -Force to skip prompt)."
		return
	}
}

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

if (-not $PSCmdlet.ShouldProcess($OutputPath, "tinygo build ($Version, $RepoSlug)")) {
	return
}

& $tinyGoExe @buildArgs
if ($LASTEXITCODE -ne 0) {
	throw "TinyGo build failed with exit code $LASTEXITCODE"
}

Get-Item $OutputPath | Select-Object -Property Name, @{l = "Version"; e = { $Version } }, @{l = "Size (KB)"; e = { $_.Length / 1KB } }, LastWriteTime, Mode, Length, Directory, FullName
