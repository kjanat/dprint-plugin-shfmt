package main

import (
	_ "embed"
	"fmt"

	"github.com/kjanat/dprint-plugin-shfmt/dprint"
)

// TinyGo's `-X` ldflag injection only takes effect on package-level
// string vars without a constant initializer. Keep `Version` etc.
// zero-valued in `main.go` and apply defaults here.
const (
	pluginName      = "dprint-plugin-shfmt"
	pluginConfigKey = "shfmt"

	defaultVersion    = "0.0.0-dev"
	defaultRepoSlug   = "kjanat/shfmt"
	defaultGitHubRepo = "kjanat/dprint-plugin-shfmt"
)

//go:generate sh -c "go-licenses report . --template licenses.tpl > licenses.generated.txt"
//go:embed licenses.generated.txt
var embeddedLicenseText string

func (h *handler) PluginInfo() dprint.PluginInfo {
	slug := orDefault(RepoSlug, defaultRepoSlug)
	updateURL := fmt.Sprintf("https://plugins.dprint.dev/%s/latest.json", slug)

	return dprint.PluginInfo{
		Name:            pluginName,
		Version:         orDefault(Version, defaultVersion),
		ConfigKey:       pluginConfigKey,
		HelpURL:         fmt.Sprintf("https://github.com/%s", orDefault(GitHubRepo, defaultGitHubRepo)),
		ConfigSchemaURL: fmt.Sprintf("https://plugins.dprint.dev/%s/%s/schema.json", slug, orDefault(ReleaseTag, defaultVersion)),
		UpdateURL:       &updateURL,
	}
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func (h *handler) LicenseText() string {
	return embeddedLicenseText
}

func (h *handler) CheckConfigUpdates(_ dprint.CheckConfigUpdatesMessage) ([]dprint.ConfigChange, error) {
	return []dprint.ConfigChange{}, nil
}
