// Package main builds the schema and wasm.
//
//	go run ./dprint/cmd/build           schema.json at root, $id=/latest/
//	go run ./dprint/cmd/build -release  .build/schema.json, $id=<git tag>
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

const (
	lockedDescription   = "Whether the configuration is not allowed to be overridden or extended."
	dprintPluginPrefix  = "dprint-plugin-"
	schemaURLFormat     = "https://plugins.dprint.dev/%s/%s/schema.json"
	schemaVersionLatest = "latest"
	tinygoStackSizeFlag = "-z stack-size=1048576"

	defaultSchemaPath = "schema.json"
	releaseSchemaPath = ".build/schema.json"
	defaultWasmPath   = "plugin.wasm"
)

type meta struct {
	ModulePath     string // github.com/kjanat/dprint-plugin-shfmt
	Host           string // github.com
	Owner          string // kjanat
	Name           string // dprint-plugin-shfmt
	Slug           string // kjanat/dprint-plugin-shfmt
	ShortSlug      string // kjanat/shfmt
	BinaryVersion  string // git describe --tags --always --dirty; for -X ldflags
	ReleaseVersion string // git describe --tags --abbrev=0 (v stripped); for schema $id in release mode
}

func main() {
	var (
		repoRoot   = flag.String("root", "", "repo root (defaults to walking up from cwd to find go.mod)")
		release    = flag.Bool("release", false, fmt.Sprintf("release mode: writes .build/schema.json with versioned $id; default writes schema.json at root with $id=/%s/", schemaVersionLatest))
		skipSchema = flag.Bool("skip-schema", false, "skip schema generation")
		skipWASM   = flag.Bool("skip-wasm", false, "skip tinygo wasm build")
		wasmOut    = flag.String("wasm-out", defaultWasmPath, "output path for the wasm binary")
	)
	flag.Parse()

	root, err := resolveRoot(*repoRoot)
	must(err)

	m, err := deriveMeta(root)
	must(err)

	schemaOut := defaultSchemaPath
	schemaVersion := schemaVersionLatest
	binaryVersion := m.BinaryVersion
	if *release {
		schemaOut = releaseSchemaPath
		schemaVersion = m.ReleaseVersion
		binaryVersion = m.ReleaseVersion
	}

	fmt.Fprintf(os.Stderr,
		"build: module=%s slug=%s short=%s binary=%s release-tag=%s mode=%s\n",
		m.ModulePath, m.Slug, m.ShortSlug, m.BinaryVersion, m.ReleaseVersion, modeLabel(*release))

	if !*skipSchema {
		must(runSchemaGen(root, m, schemaOut, schemaVersion))
	}
	if !*skipWASM {
		must(runTinygo(root, m, *wasmOut, binaryVersion))
	}
}

func modeLabel(release bool) string {
	if release {
		return "release"
	}
	return "local"
}

func resolveRoot(override string) (string, error) {
	if override != "" {
		abs, err := filepath.Abs(override)
		return abs, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found walking up from %s", cwd)
		}
		dir = parent
	}
}

func deriveMeta(root string) (meta, error) {
	modBytes, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return meta{}, fmt.Errorf("read go.mod: %w", err)
	}
	mp := modfile.ModulePath(modBytes)
	if mp == "" {
		return meta{}, fmt.Errorf("go.mod has no module path")
	}

	parts := strings.Split(mp, "/")
	if len(parts) < 3 {
		return meta{}, fmt.Errorf("module path %q has fewer than 3 segments", mp)
	}
	host, owner, name := parts[0], parts[1], parts[2]
	short := strings.TrimPrefix(name, dprintPluginPrefix)

	binaryVersion, err := gitOutput(root, "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return meta{}, fmt.Errorf("git describe --dirty: %w", err)
	}
	releaseTag, err := gitOutput(root, "describe", "--tags", "--abbrev=0")
	if err != nil {
		return meta{}, fmt.Errorf("git describe --abbrev=0: %w", err)
	}

	return meta{
		ModulePath:     mp,
		Host:           host,
		Owner:          owner,
		Name:           name,
		Slug:           owner + "/" + name,
		ShortSlug:      owner + "/" + short,
		BinaryVersion:  binaryVersion,
		ReleaseVersion: strings.TrimPrefix(releaseTag, "v"),
	}, nil
}

func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func runSchemaGen(root string, m meta, outPath, schemaVersion string) error {
	if dir := filepath.Dir(outPath); dir != "" && dir != "." {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	schemaID := fmt.Sprintf(schemaURLFormat, m.ShortSlug, schemaVersion)
	args := []string{
		"run", "./dprint/cmd/gen-json-schema",
		"-type", "configuration",
		"-out", outPath,
		"-schema-id", schemaID,
		"-include-locked",
		"-locked-description", lockedDescription,
	}
	return runIn(root, "go", args...)
}

func runTinygo(root string, m meta, outPath, binaryVersion string) error {
	ldflags := fmt.Sprintf(
		"-extldflags '%s' -X main.Version=%s -X main.ReleaseTag=%s -X main.RepoSlug=%s -X main.GitHubRepo=%s",
		tinygoStackSizeFlag, binaryVersion, binaryVersion, m.ShortSlug, m.Slug,
	)
	args := []string{
		"build",
		"-o", outPath,
		"-target=wasm-unknown",
		"-gc=conservative",
		"-scheduler=none",
		"-panic=trap",
		"-no-debug",
		"-ldflags=" + ldflags,
		".",
	}
	return runIn(root, "tinygo", args...)
}

func runIn(dir, bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "build:", err)
		os.Exit(1)
	}
}
