package main

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/kjanat/dprint-plugin-shfmt/dprint"
)

func TestResolveConfigDefaults(t *testing.T) {
	h := &handler{}

	result := h.ResolveConfig(dprint.ConfigKeyMap{}, dprint.GlobalConfiguration{})

	if result.Config.IndentWidth != 2 {
		t.Fatalf("expected default indent width 2, got %d", result.Config.IndentWidth)
	}
	if result.Config.UseTabs {
		t.Fatal("expected default useTabs to be false")
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(result.Diagnostics))
	}
	for _, fileExtension := range result.FileMatching.FileExtensions {
		if fileExtension == extZsh {
			t.Fatal("expected zsh not to be advertised by default")
		}
	}
}

func TestResolveConfigExperimentalZshAdvertisesExtension(t *testing.T) {
	h := &handler{}

	result := h.ResolveConfig(
		dprint.ConfigKeyMap{cfgKeyExperimentalZsh: true},
		dprint.GlobalConfiguration{},
	)

	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(result.Diagnostics))
	}
	found := slices.Contains(result.FileMatching.FileExtensions, extZsh)
	if !found {
		t.Fatal("expected zsh to be advertised when experimentalZsh is true")
	}
}

func TestResolveConfigAllowsLockedProperty(t *testing.T) {
	h := &handler{}

	result := h.ResolveConfig(
		dprint.ConfigKeyMap{
			"locked": true,
		},
		dprint.GlobalConfiguration{},
	)

	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics for locked property, got %d", len(result.Diagnostics))
	}
}

func TestResolveConfigPluginConfigPrecedenceAndDiagnostics(t *testing.T) {
	h := &handler{}

	result := h.ResolveConfig(
		dprint.ConfigKeyMap{
			cfgKeyIndentWidth:  float64(4),
			cfgKeyUseTabs:      false,
			cfgKeyFuncNextLine: "invalid",
			"unknownField":     true,
		},
		dprint.GlobalConfiguration{
			cfgKeyIndentWidth: float64(8),
			cfgKeyUseTabs:     true,
		},
	)

	if result.Config.IndentWidth != 4 {
		t.Fatalf("expected plugin indent width to take precedence, got %d", result.Config.IndentWidth)
	}
	if result.Config.UseTabs {
		t.Fatal("expected plugin useTabs to take precedence")
	}

	if len(result.Diagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(result.Diagnostics))
	}
}

func TestResolveConfigCoercesFlexibleValueTypes(t *testing.T) {
	h := &handler{}

	result := h.ResolveConfig(
		dprint.ConfigKeyMap{
			cfgKeyIndentWidth:      []byte("4"),
			cfgKeyUseTabs:          "false",
			cfgKeyBinaryNextLine:   []byte("true"),
			cfgKeySwitchCaseIndent: json.Number("1"),
			cfgKeySpaceRedirects:   float64(0),
		},
		dprint.GlobalConfiguration{
			cfgKeyIndentWidth: json.Number("8"),
			cfgKeyUseTabs:     []byte("1"),
		},
	)

	if result.Config.IndentWidth != 4 {
		t.Fatalf("expected coerced plugin indent width 4, got %d", result.Config.IndentWidth)
	}
	if result.Config.UseTabs {
		t.Fatal("expected coerced plugin useTabs to be false")
	}
	if !result.Config.BinaryNextLine {
		t.Fatal("expected binaryNextLine to be true")
	}
	if !result.Config.SwitchCaseIndent {
		t.Fatal("expected switchCaseIndent to be true")
	}
	if result.Config.SpaceRedirects {
		t.Fatal("expected spaceRedirects to be false")
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(result.Diagnostics))
	}
}

func TestResolveConfigIgnoresNilValues(t *testing.T) {
	h := &handler{}

	result := h.ResolveConfig(
		dprint.ConfigKeyMap{
			cfgKeyIndentWidth: nil,
			cfgKeyUseTabs:     nil,
		},
		dprint.GlobalConfiguration{
			cfgKeyIndentWidth: nil,
			cfgKeyUseTabs:     nil,
		},
	)

	if result.Config.IndentWidth != 2 {
		t.Fatalf("expected fallback indent width 2, got %d", result.Config.IndentWidth)
	}
	if result.Config.UseTabs {
		t.Fatal("expected fallback useTabs false")
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics for nil values, got %d", len(result.Diagnostics))
	}
}

func TestFormatWithShfmt(t *testing.T) {
	h := &handler{}

	result := h.Format(
		dprint.SyncFormatRequest[configuration]{
			FilePath: "sample.sh",
			FileBytes: []byte(
				"if [ \"$1\" = \"ok\" ];then\n echo ok\nfi\n",
			),
			Config: configuration{
				IndentWidth: 2,
				UseTabs:     false,
			},
		},
		nil,
	)

	if result.Code != dprint.FormatResultChange {
		t.Fatalf("expected format result code %d, got %d", dprint.FormatResultChange, result.Code)
	}

	expected := "if [ \"$1\" = \"ok\" ]; then\n  echo ok\nfi\n"
	if string(result.Text) != expected {
		t.Fatalf("unexpected formatted output:\n%s", string(result.Text))
	}
}

func TestFormatAppliesUpstreamSyntaxBehavior(t *testing.T) {
	cases := []struct {
		name     string
		filePath string
		input    string
		expected string
	}{
		{
			name:     "space after arithmetic negation",
			filePath: testFileScriptBash,
			input:    "if ((!count)); then\n  echo empty\nfi\n",
			expected: "if ((! count)); then\n  echo empty\nfi\n",
		},
		{
			name:     "array element in brace redirect",
			filePath: testFileScriptBash,
			input:    "exec {fds[3]}>&-\n",
			expected: "exec {fds[3]}>&-\n",
		},
		{
			name:     "pound sign in associative array key",
			filePath: testFileScriptBash,
			input:    "echo \"${args[cmd,#]}\"\n",
			expected: "echo \"${args[cmd,#]}\"\n",
		},
	}

	h := &handler{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Format(
				dprint.SyncFormatRequest[configuration]{
					FilePath:  tc.filePath,
					FileBytes: []byte(tc.input),
					Config: configuration{
						IndentWidth: 2,
						UseTabs:     false,
					},
				},
				nil,
			)

			if result.Code == dprint.FormatResultError {
				t.Fatalf("expected input to parse, got error: %v", result.Err)
			}

			formatted := tc.input
			if result.Code == dprint.FormatResultChange {
				formatted = string(result.Text)
			}
			if formatted != tc.expected {
				t.Fatalf("unexpected formatted output:\n%s", formatted)
			}
		})
	}
}

func TestFormatDetectsBashShebang(t *testing.T) {
	h := &handler{}

	result := h.Format(
		dprint.SyncFormatRequest[configuration]{
			FilePath: testFileScriptSh,
			FileBytes: []byte(
				"#!/usr/bin/env bash\nif [[ \"$a\" == \"b\" ]];then\necho ok\nfi\n",
			),
			Config: configuration{
				IndentWidth: 2,
				UseTabs:     false,
			},
		},
		nil,
	)

	if result.Code == dprint.FormatResultError {
		t.Fatalf("expected bash shebang to parse, got error: %v", result.Err)
	}
}

func TestFormatReturnsErrorOnParseFailure(t *testing.T) {
	h := &handler{}

	result := h.Format(
		dprint.SyncFormatRequest[configuration]{
			FilePath:  "broken.sh",
			FileBytes: []byte("if [ \"$1\" = \"ok\" ]; then\n"),
			Config: configuration{
				IndentWidth: 2,
				UseTabs:     false,
			},
		},
		nil,
	)

	if result.Code != dprint.FormatResultError {
		t.Fatalf("expected format error code %d, got %d", dprint.FormatResultError, result.Code)
	}
	if result.Err == nil || !strings.Contains(result.Err.Error(), "then") {
		t.Fatalf("unexpected error text: %v", result.Err)
	}
}

func TestLicenseTextEmbedsFullLicenseReport(t *testing.T) {
	h := &handler{}
	licenseText := h.LicenseText()

	if strings.TrimSpace(licenseText) == "" {
		t.Fatal("expected non-empty license text")
	}
	if !strings.Contains(licenseText, "## github.com/kjanat/dprint-plugin-shfmt") {
		t.Fatal("expected first-party module section in embedded license text")
	}
	if !strings.Contains(licenseText, "BSD 3-Clause License") {
		t.Fatal("expected full BSD license text in embedded license text")
	}
}

func withInjectedMetadata(t *testing.T, version, tag, slug, ghRepo string) {
	t.Helper()
	oldV, oldT, oldS, oldG := Version, ReleaseTag, RepoSlug, GitHubRepo
	Version, ReleaseTag, RepoSlug, GitHubRepo = version, tag, slug, ghRepo
	t.Cleanup(func() {
		Version, ReleaseTag, RepoSlug, GitHubRepo = oldV, oldT, oldS, oldG
	})
}

func TestPluginInfoUsesUpstreamDefaults(t *testing.T) {
	withInjectedMetadata(t, "1.2.3", "1.2.3", "hrko/shfmt", "hrko/dprint-plugin-shfmt")

	info := (&handler{}).PluginInfo()

	if info.UpdateURL == nil {
		t.Fatal("expected update URL to be set")
	}
	if *info.UpdateURL != "https://plugins.dprint.dev/hrko/shfmt/latest.json" {
		t.Fatalf("unexpected update URL: %q", *info.UpdateURL)
	}
	if info.ConfigSchemaURL != "https://plugins.dprint.dev/hrko/shfmt/1.2.3/schema.json" {
		t.Fatalf("unexpected config schema URL: %q", info.ConfigSchemaURL)
	}
	if info.HelpURL != "https://github.com/hrko/dprint-plugin-shfmt" {
		t.Fatalf("unexpected help URL: %q", info.HelpURL)
	}
}

func TestPluginInfoUsesInjectedForkMetadata(t *testing.T) {
	withInjectedMetadata(t, "0.0.5", "0.0.5", "kjanat/shfmt", "kjanat/dprint-plugin-shfmt")

	info := (&handler{}).PluginInfo()

	if info.UpdateURL == nil {
		t.Fatal("expected update URL to be set")
	}
	if *info.UpdateURL != "https://plugins.dprint.dev/kjanat/shfmt/latest.json" {
		t.Fatalf("unexpected update URL: %q", *info.UpdateURL)
	}
	if info.ConfigSchemaURL != "https://plugins.dprint.dev/kjanat/shfmt/0.0.5/schema.json" {
		t.Fatalf("unexpected config schema URL: %q", info.ConfigSchemaURL)
	}
	if info.HelpURL != "https://github.com/kjanat/dprint-plugin-shfmt" {
		t.Fatalf("unexpected help URL: %q", info.HelpURL)
	}
}
