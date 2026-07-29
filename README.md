# haunt

Your commands haunt the directories they were run in.

Haunt is a directory-aware shell history tool. It records every command alongside its working directory, then lets you search history scoped to where you are — so the commands you see are the ones you actually ran *here*, not noise from every other project.

Global history (Ctrl+R) stays untouched. Haunt adds a separate keybinding (Ctrl+G by default) for directory-scoped search.

## Install

```sh
go install github.com/alde/haunt@latest
```

Optionally install [fzf](https://github.com/junegunn/fzf) for interactive fuzzy search. Without it, haunt falls back to printing matching commands to stdout.

## Shell setup

### Fish

Add to `~/.config/fish/config.fish`:

```fish
haunt init fish | source
```

### Zsh

Add to `~/.zshrc`:

```zsh
eval "$(haunt init zsh)"
```

#### Cached eval (faster shell startup)

`haunt init zsh` output only changes when the `haunt` binary is updated. You can cache it to avoid the subprocess on every shell start:

```zsh
# Cache eval output from slow tools — regenerates when the binary updates
_cached_eval() {
    local cache_dir="$HOME/.zsh/cache"
    local cmd_name="${1##*/}"
    local cache_file="$cache_dir/$cmd_name"
    local bin_path="${1}"

    [[ -x "$bin_path" ]] || return 0
    if [[ ! -f "$cache_file" ]] || [[ "$bin_path" -nt "$cache_file" ]]; then
        mkdir -p "$cache_dir"
        "${@}" > "$cache_file" 2>/dev/null
    fi
    source "$cache_file"
}

_cached_eval "$(command -v haunt)" init zsh
```

This sources a cached copy from `~/.zsh/cache/haunt` and only regenerates it when the binary's mtime changes (e.g. after `go install`). The `_cached_eval` function works for any tool that outputs shell code — just call it the same way for other `eval "$(… init)"` patterns.

## Usage

Once the shell integration is loaded, commands are recorded automatically. Press **Ctrl+G** to search history scoped to your current directory.

```
haunt search    # interactive directory-scoped search (called by the keybinding)
haunt config    # show current configuration
haunt init fish # print fish integration script
haunt init zsh  # print zsh integration script
```

## Scope modes

Haunt supports three scoping strategies, configured via `scope` in the config file:

| Mode | Behaviour |
|------|-----------|
| `ancestors` (default) | Commands from the current directory and all parent directories up to the git root. A command run at the repo root shows up in every subdirectory. |
| `git-root` | Commands from anywhere within the same git repository. |
| `exact` | Only commands from the exact current directory. |

## Configuration

Config lives at `~/.config/haunt/config.toml` (respects `XDG_CONFIG_HOME`):

```toml
keybinding = "ctrl-g"    # also supports alt-r, ctrl-alt-r, etc.
scope = "ancestors"      # ancestors, git-root, or exact
db_path = "~/.local/share/haunt/history.db"
```

All fields are optional — sensible defaults are used for anything missing.

## How it works

1. A shell hook records every command with its working directory and timestamp into a local bbolt database.
2. When you press the keybinding, haunt queries the database using the current scope mode and pipes results through fzf.
3. The selected command is placed on your command line, ready to run or edit.

Recording runs in the background and won't slow down your shell.
