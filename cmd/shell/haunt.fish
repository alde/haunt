function __haunt_record --on-event fish_postexec
    set -l last_status $status
    haunt record --exit-code $last_status -- $argv &
    disown
end

function __haunt_search
    set -l result (haunt search)
    if test -n "$result"
        commandline -r -- $result
        commandline -f repaint
    end
end

bind {{KEYBINDING}} __haunt_search
