__haunt_record() {
    haunt record --exit-code $? -- "$1" &!
}
autoload -Uz add-zsh-hook
add-zsh-hook zshaddhistory __haunt_record

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
