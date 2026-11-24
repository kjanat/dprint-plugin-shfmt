package main

import (
	"encoding/json"
	"maps"

	"github.com/kjanat/dprint-plugin-shfmt/dprint"
)

//go:generate go run github.com/kjanat/dprint-plugin-shfmt/dprint/cmd/gen-config-resolver -type configuration -out handler_config_generated.go -extra-known-keys locked
//go:generate go run github.com/kjanat/dprint-plugin-shfmt/dprint/cmd/gen-json-schema -type configuration -out schema.json -schema-id https://plugins.dprint.dev/kjanat/shfmt/latest/schema.json -include-locked -locked-description "Whether the configuration is not allowed to be overridden or extended."

type configuration struct {
	IndentWidth      uint32 `description:"Number of spaces per indentation level when not using tabs."                                        dprint:"default=2,global"     json:"indentWidth"`
	UseTabs          bool   `description:"Whether to use tabs for indentation."                                                               dprint:"default=false,global" json:"useTabs"`
	BinaryNextLine   bool   `description:"Whether binary operators should be placed at the start of the next line when line wrapping occurs." dprint:"default=false"        json:"binaryNextLine"`
	SwitchCaseIndent bool   `description:"Whether switch case bodies should be indented."                                                     dprint:"default=false"        json:"switchCaseIndent"`
	SpaceRedirects   bool   `description:"Whether to insert a space after redirection operators."                                             dprint:"default=false"        json:"spaceRedirects"`
	FuncNextLine     bool   `description:"Whether to place function opening braces on the next line."                                         dprint:"default=false"        json:"funcNextLine"`
	Minify           bool   `description:"Whether to minify shell scripts when printing."                                                     dprint:"default=false"        json:"minify"`
	ExperimentalZsh  bool   `description:"Whether to format Zsh scripts, by .zsh extension or zsh shebang."                                   dprint:"default=true"         json:"experimentalZsh"`
}

const configFilePayloadKey = "$configFile"

var fileExtensions = []string{extSh, extBash, extMksh, extBats}

func (h *handler) ResolveConfig(
	config dprint.ConfigKeyMap,
	global dprint.GlobalConfiguration,
) dprint.ResolveConfigurationResult[configuration] {
	merged, fileDiagnostics := configWithFile(config)
	resolved, diagnostics := dprint.ResolveConfigWithSpec(
		merged,
		global,
		generatedConfigurationResolverSpec,
	)

	extensions := append([]string(nil), fileExtensions...)
	if resolved.ExperimentalZsh {
		extensions = append(extensions, extZsh)
	}

	return dprint.ResolveConfigurationResult[configuration]{
		FileMatching: dprint.FileMatchingInfo{
			FileExtensions: extensions,
			FileNames:      []string{},
		},
		Diagnostics: append(fileDiagnostics, diagnostics...),
		Config:      resolved,
	}
}

// configWithFile combines kprint's parsed file content with inline settings.
// The core Wasm guest never opens the file itself.
func configWithFile(inline dprint.ConfigKeyMap) (dprint.ConfigKeyMap, []dprint.ConfigurationDiagnostic) {
	payload, present := inline[configFilePayloadKey]
	if !present {
		return inline, nil
	}

	overrides := maps.Clone(inline)
	delete(overrides, configFilePayloadKey)
	envelope, _ := payload.(map[string]any)
	content, _ := envelope["content"].(string)

	var file dprint.ConfigKeyMap
	if err := json.Unmarshal([]byte(content), &file); err != nil || file == nil {
		return overrides, []dprint.ConfigurationDiagnostic{{
			"propertyName": "configFile",
			"message":      "Expected the selected shfmt configuration to be a JSON object.",
		}}
	}

	maps.Copy(file, overrides)
	return file, nil
}
