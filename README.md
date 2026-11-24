# dprint-plugin-shfmt

Shell script formatting plugin for dprint.

This uses the [`mvdan.cc/sh/v3`](https://github.com/mvdan/sh) parser and printer used by `shfmt`.

> Fork of [`hrko/dprint-plugin-shfmt`](https://github.com/hrko/dprint-plugin-shfmt). BSD-3-Clause; upstream copyright preserved in [LICENSE](./LICENSE).

## Setup

You can add this plugin to your dprint config with:

```sh
dprint add kjanat/shfmt
```

## Example config

This example enables the plugin, targets shell script files, and sets a few common formatting options.
`indentWidth` and `useTabs` are global dprint options, while settings under `shfmt` are plugin-specific.
When both global and plugin values are set for the same option, the plugin value takes precedence.

```json
{
  "plugins": ["https://plugins.dprint.dev/kjanat/shfmt-<version>.wasm"],
  "includes": ["**/*.sh", "**/*.bash"],
  "indentWidth": 2,
  "useTabs": false,
  "shfmt": {
    "switchCaseIndent": true,
    "spaceRedirects": true,
    "funcNextLine": false
  }
}
```

> [!NOTE]
> Zsh formatting is on by default, but support in the underlying library is still experimental ([mvdan/sh `zsh` label](https://github.com/mvdan/sh/labels/zsh)).\
> To opt out, set `"shfmt": {"experimentalZsh": false}` in your dprint config.\
> Once opted out, `.zsh` files are not claimed by the plugin and a zsh shebang in any other file is left unformatted.

### Shared configuration with kprint

With kprint, set `shfmt.configFile` to a JSON, YAML, or TOML file. Use a fragment
to select a table from a shared file: `"configFile": "formatters.toml#tool.shfmt"`
loads the `[tool.shfmt]` table. Use the same option names as in the inline
`shfmt` object. Inline plugin options take precedence over file options, which
take precedence over the global defaults. The plugin does not look for
configuration files on its own. Use inline options with upstream dprint.

## Configuration schema

See the schema for all available options and the latest canonical definitions.

- [schema.json](./schema.json)

## Development docs

For development documentation, see [AGENTS.md](./AGENTS.md).
