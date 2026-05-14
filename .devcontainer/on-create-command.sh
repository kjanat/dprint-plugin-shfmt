#!/bin/bash

mkdir -p ~/.local/bin
mkdir -p ~/.config/fish

append_once() {
  local line=$1 file=$2
  grep -qxF -- "$line" "$file" 2>/dev/null || echo "$line" >>"$file"
}

curl https://mise.run | sh
append_once 'eval "$(~/.local/bin/mise activate bash)"' ~/.bashrc
append_once 'eval "$(~/.local/bin/mise activate zsh)"' ~/.zshrc
append_once '~/.local/bin/mise activate fish | source' ~/.config/fish/config.fish

curl -sS https://starship.rs/install.sh | env BIN_DIR=~/.local/bin FORCE=1 sh
append_once 'eval "$(starship init bash)"' ~/.bashrc
append_once 'eval "$(starship init zsh)"' ~/.zshrc
append_once 'starship init fish | source' ~/.config/fish/config.fish

curl --proto '=https' --tlsv1.2 -LsSf https://setup.atuin.sh | env ATUIN_INSTALL_DIR=~/.local/bin sh
# Atuin will automatically update the shell configuration files, so no additional setup is needed here.
