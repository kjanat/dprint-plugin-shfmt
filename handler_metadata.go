package main

import (
	_ "embed"
	"fmt"

	"github.com/hrko/dprint-plugin-shfmt/dprint"
)

const (
	pluginName      = "dprint-plugin-shfmt"
	pluginConfigKey = "shfmt"
)

//go:generate sh -c "go-licenses report . --template licenses.tpl > licenses.generated.txt"
//go:embed licenses.generated.txt
var embeddedLicenseText string

func (h *handler) PluginInfo() dprint.PluginInfo {
	updateURL := fmt.Sprintf("https://plugins.dprint.dev/%s/latest.json", RepoSlug)

	return dprint.PluginInfo{
		Name:            pluginName,
		Version:         Version,
		ConfigKey:       pluginConfigKey,
		HelpURL:         fmt.Sprintf("https://github.com/%s", GitHubRepo),
		ConfigSchemaURL: fmt.Sprintf("https://plugins.dprint.dev/%s/%s/schema.json", RepoSlug, ReleaseTag),
		UpdateURL:       &updateURL,
	}
}

func (h *handler) LicenseText() string {
	return embeddedLicenseText
}

func (h *handler) CheckConfigUpdates(_ dprint.CheckConfigUpdatesMessage) ([]dprint.ConfigChange, error) {
	return []dprint.ConfigChange{}, nil
}
