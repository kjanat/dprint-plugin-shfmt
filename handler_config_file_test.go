package main

import (
	"testing"

	"github.com/kjanat/dprint-plugin-shfmt/dprint"
)

func TestMalformedConfigFileReportsConfigDiagnostic(t *testing.T) {
	h := &handler{}
	for _, content := range []string{"", "null", "[]"} {
		config := dprint.ConfigKeyMap{
			configFilePayloadKey: map[string]any{"content": content},
			cfgKeyIndentWidth:    6,
		}
		result := h.ResolveConfig(config, nil)
		if len(result.Diagnostics) != 1 || result.Diagnostics[0]["propertyName"] != "configFile" {
			t.Errorf("content %q: expected a configFile diagnostic, got %v", content, result.Diagnostics)
		}
		if result.Config.IndentWidth != 6 {
			t.Errorf("content %q: expected inline indent width 6, got %d", content, result.Config.IndentWidth)
		}
	}
}
