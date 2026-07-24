__haunt_preexec() {
    _haunt_cmd="$1"
}

__haunt_precmd() {
    local exit_code=$?
    [[ -n "$_haunt_cmd" ]] || return
    haunt record --exit-code $exit_code -- "$_haunt_cmd" &!
    _haunt_cmd=""
}

autoload -Uz add-zsh-hook
add-zsh-hook preexec __haunt_preexec
add-zsh-hook precmd __haunt_precmd

__haunt_search() {
    local result
    result=$(haunt search)
    if [[ -n "$result" ]]; then
        BUFFER="$result"
        CURSOR=${#BUFFER}
        zle redisplay
    fi
}
zle -N __haunt_search

bindkey '{{KEYBINDING}}' __haunt_search
