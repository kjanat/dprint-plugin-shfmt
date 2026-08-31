package main

import (
	"bytes"

	"mvdan.cc/sh/v3/syntax"

	"github.com/kjanat/dprint-plugin-shfmt/dprint"
)

func (h *handler) Format(
	request dprint.SyncFormatRequest[configuration],
	_ dprint.HostFormatFunc,
) dprint.FormatResult {
	variant := detectVariant(request.FilePath, request.FileBytes)

	// The .zsh extension is only advertised while experimentalZsh is set, but a
	// zsh shebang reaches this handler through any advertised extension. Honour
	// the opt-out for both routes and leave such files untouched.
	if variant == syntax.LangZsh && !request.Config.ExperimentalZsh {
		return dprint.NoChange()
	}

	parser := syntax.NewParser(
		syntax.Variant(variant),
		syntax.KeepComments(true),
	)
	prog, err := parser.Parse(bytes.NewReader(request.FileBytes), request.FilePath)
	if err != nil {
		return dprint.FormatError(err)
	}

	printer := syntax.NewPrinter(
		syntax.Indent(indentSize(request.Config)),
		syntax.BinaryNextLine(request.Config.BinaryNextLine),
		syntax.SwitchCaseIndent(request.Config.SwitchCaseIndent),
		syntax.SpaceRedirects(request.Config.SpaceRedirects),
		syntax.FunctionNextLine(request.Config.FuncNextLine),
		syntax.Minify(request.Config.Minify),
	)

	var buffer bytes.Buffer
	if err := printer.Print(&buffer, prog); err != nil {
		return dprint.FormatError(err)
	}

	formatted := buffer.Bytes()
	if bytes.Equal(request.FileBytes, formatted) {
		return dprint.NoChange()
	}

	return dprint.Change(append([]byte(nil), formatted...))
}

func indentSize(config configuration) uint {
	if config.UseTabs {
		return 0
	}
	return uint(config.IndentWidth)
}
