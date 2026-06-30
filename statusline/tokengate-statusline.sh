#!/bin/bash
# TokenGate Status Line for Claude Code
#
# Default mode:
#   TokenGate-first cost and budget status.
#
# Compatibility mode:
#   Current Claude Code-oriented status line behavior from the burnrate-ai
#   prototype, including Claude OAuth 5h/7d usage windows.
#
# Setup:
#   { "statusLine": { "type": "command", "command": "sh ~/.claude/tokengate-statusline.sh" } }
#
# Mode selection:
#   sh ~/.claude/tokengate-statusline.sh --mode tokengate
#   sh ~/.claude/tokengate-statusline.sh --mode claude
#   TOKENGATE_STATUSLINE_MODE=tokengate|claude

set -f

statusline_mode="${TOKENGATE_STATUSLINE_MODE:-tokengate}"
while [ "$#" -gt 0 ]; do
    case "$1" in
        --mode)
            statusline_mode="${2:-tokengate}"
            shift 2
            ;;
        --mode=*)
            statusline_mode="${1#--mode=}"
            shift
            ;;
        *)
            shift
            ;;
    esac
done

case "$statusline_mode" in
    tokengate|claude) ;;
    *) statusline_mode="tokengate" ;;
esac

input=$(cat)

if [ -z "$input" ]; then
    printf "Claude"
    exit 0
fi

# ANSI colors
blue='\033[38;2;0;153;255m'
orange='\033[38;2;255;176;85m'
green='\033[38;2;0;160;0m'
cyan='\033[38;2;46;149;153m'
red='\033[38;2;255;85;85m'
yellow='\033[38;2;230;200;0m'
white='\033[38;2;220;220;220m'
magenta='\033[38;2;200;120;255m'
dim='\033[2m'
reset='\033[0m'

cache_dir="${TOKENGATE_STATUSLINE_CACHE_DIR:-/tmp/claude}"
mkdir -p "$cache_dir" 2>/dev/null || true

format_tokens() {
    local num=${1:-0}
    if [ "$num" -ge 1000000 ] 2>/dev/null; then
        awk "BEGIN {printf \"%.1fm\", $num / 1000000}"
    elif [ "$num" -ge 1000 ] 2>/dev/null; then
        awk "BEGIN {printf \"%.0fk\", $num / 1000}"
    else
        printf "%d" "$num"
    fi
}

format_money() {
    local amount=${1:-0}
    awk "BEGIN {
        value = $amount + 0
        if (value == int(value)) {
            printf \"\$%.0f\", value
        } else {
            printf \"\$%.2f\", value
        }
    }"
}

usage_color() {
    local pct=${1:-0}
    if [ "$pct" -ge 90 ] 2>/dev/null; then echo "$red"
    elif [ "$pct" -ge 70 ] 2>/dev/null; then echo "$orange"
    elif [ "$pct" -ge 50 ] 2>/dev/null; then echo "$yellow"
    else echo "$green"
    fi
}

dot_bar() {
    local pct=${1:-0}
    local blocks=${2:-6}
    local filled=$(( (pct * blocks + 99) / 100 ))
    [ "$pct" -eq 0 ] 2>/dev/null && filled=0
    [ "$filled" -gt "$blocks" ] && filled=$blocks
    [ "$filled" -lt 0 ] && filled=0
    local empty=$(( blocks - filled ))
    local bar=""
    local i
    for (( i=0; i<filled; i++ )); do bar+="●"; done
    for (( i=0; i<empty; i++ )); do bar+="○"; done
    echo "$bar"
}

visible_len() {
    printf "%b" "$1" | sed $'s/\033\[[0-9;]*m//g' | wc -m | tr -d ' '
}

iso_to_epoch() {
    local iso_str="$1"
    local epoch
    epoch=$(date -d "${iso_str}" +%s 2>/dev/null)
    if [ -n "$epoch" ]; then echo "$epoch"; return 0; fi
    local stripped="${iso_str%%.*}"
    stripped="${stripped%%Z}"
    stripped="${stripped%%+*}"
    stripped="${stripped%%-[0-9][0-9]:[0-9][0-9]}"
    if [[ "$iso_str" == *"Z"* ]] || [[ "$iso_str" == *"+00:00"* ]] || [[ "$iso_str" == *"-00:00"* ]]; then
        epoch=$(env TZ=UTC date -j -f "%Y-%m-%dT%H:%M:%S" "$stripped" +%s 2>/dev/null)
    else
        epoch=$(date -j -f "%Y-%m-%dT%H:%M:%S" "$stripped" +%s 2>/dev/null)
    fi
    if [ -n "$epoch" ]; then echo "$epoch"; return 0; fi
    return 1
}

format_reset_time() {
    local iso_str="$1"
    local style="$2"
    [ -z "$iso_str" ] || [ "$iso_str" = "null" ] && return
    local epoch
    epoch=$(iso_to_epoch "$iso_str")
    [ -z "$epoch" ] && return
    case "$style" in
        time)
            date -j -r "$epoch" +"%l:%M%P" 2>/dev/null | sed 's/^ //' || \
            date -d "@$epoch" +"%l:%M%P" 2>/dev/null | sed 's/^ //'
            ;;
        datetime)
            date -j -r "$epoch" +"%b %-d, %l:%M%P" 2>/dev/null | sed 's/  / /g; s/^ //' || \
            date -d "@$epoch" +"%b %-d, %l:%M%P" 2>/dev/null | sed 's/  / /g; s/^ //'
            ;;
    esac
}

get_oauth_token() {
    if [ -n "$CLAUDE_CODE_OAUTH_TOKEN" ]; then
        echo "$CLAUDE_CODE_OAUTH_TOKEN"; return 0
    fi
    if command -v security >/dev/null 2>&1; then
        local blob
        blob=$(security find-generic-password -s "Claude Code-credentials" -w 2>/dev/null)
        if [ -n "$blob" ]; then
            local token
            token=$(echo "$blob" | jq -r '.claudeAiOauth.accessToken // empty' 2>/dev/null)
            if [ -n "$token" ] && [ "$token" != "null" ]; then echo "$token"; return 0; fi
        fi
    fi
    local creds_file="${HOME}/.claude/.credentials.json"
    if [ -f "$creds_file" ]; then
        local token
        token=$(jq -r '.claudeAiOauth.accessToken // empty' "$creds_file" 2>/dev/null)
        if [ -n "$token" ] && [ "$token" != "null" ]; then echo "$token"; return 0; fi
    fi
    if command -v secret-tool >/dev/null 2>&1; then
        local blob
        blob=$(timeout 2 secret-tool lookup service "Claude Code-credentials" 2>/dev/null)
        if [ -n "$blob" ]; then
            local token
            token=$(echo "$blob" | jq -r '.claudeAiOauth.accessToken // empty' 2>/dev/null)
            if [ -n "$token" ] && [ "$token" != "null" ]; then echo "$token"; return 0; fi
        fi
    fi
    echo ""
}

read_json() {
    echo "$input" | jq -r "$1" 2>/dev/null
}

model_name=$(read_json '.model.display_name // "Claude"')
size=$(read_json '.context_window.context_window_size // 200000')
[ "$size" -eq 0 ] 2>/dev/null && size=200000

input_tokens=$(read_json '.context_window.current_usage.input_tokens // 0')
cache_create=$(read_json '.context_window.current_usage.cache_creation_input_tokens // 0')
cache_read=$(read_json '.context_window.current_usage.cache_read_input_tokens // 0')
current=$(( input_tokens + cache_create + cache_read ))

used_tokens=$(format_tokens "$current")
total_tokens=$(format_tokens "$size")

ctx_pct=0
ctx_remain=100
if [ "$size" -gt 0 ] 2>/dev/null; then
    ctx_pct=$(( current * 100 / size ))
    ctx_remain=$(( 100 - ctx_pct ))
fi
ctx_color=$(usage_color "$ctx_pct")

session_cost=$(read_json '.cost // 0' | LC_NUMERIC=C awk '{printf "%.4f", $1}')
session_cost_nonzero=false
if LC_NUMERIC=C awk "BEGIN {exit !($session_cost > 0.00005)}"; then
    session_cost_nonzero=true
fi

cwd=$(read_json '.cwd // empty')
display_dir=""
if [ -n "$cwd" ]; then
    display_dir="${cwd##*/}"
    git_branch=$(git -C "$cwd" rev-parse --abbrev-ref HEAD 2>/dev/null)
    if [ -n "$git_branch" ]; then
        display_dir="${display_dir}@${git_branch}"
    fi
fi

tg_key="${TOKENGATE_API_KEY:-${ANTHROPIC_AUTH_TOKEN:-${ANTHROPIC_API_KEY:-}}}"
if [ -z "$tg_key" ] && [ -n "$ANTHROPIC_CUSTOM_HEADERS" ]; then
    tg_key=$(printf '%s' "$ANTHROPIC_CUSTOM_HEADERS" | grep -oE 'X-TokenGate-Key:[^,[:space:]]+' | head -1 | cut -d: -f2-)
fi

tg_base="${ANTHROPIC_BASE_URL:-}"
tg_base="${tg_base%/}"
tg_poll="${TOKENGATE_STATUSLINE_POLL:-60}"
tg_blocks="${TOKENGATE_STATUSLINE_BARS:-6}"
cols=${COLUMNS:-200}

fetch_cached_json() {
    local file="$1"
    local max_age="$2"
    if [ ! -f "$file" ]; then
        return 1
    fi
    local cache_mtime
    cache_mtime=$(stat -c %Y "$file" 2>/dev/null || stat -f %m "$file" 2>/dev/null)
    local now_epoch
    now_epoch=$(date +%s)
    local cache_age=$(( now_epoch - cache_mtime ))
    if [ "$cache_age" -lt "$max_age" ]; then
        cat "$file" 2>/dev/null
        return 0
    fi
    return 1
}

fetch_url_json() {
    local url="$1"
    curl -s --max-time 2 \
        -H "X-TokenGate-Key: ${tg_key}" \
        -H "Authorization: Bearer ${tg_key}" \
        -H "Accept: application/json" \
        "$url" 2>/dev/null
}

fetch_tokengate_data() {
    local cache_file="$cache_dir/tokengate-statusline-cache.json"
    local data
    data=$(fetch_cached_json "$cache_file" "$tg_poll")
    if [ -n "$data" ]; then
        echo "$data"
        return 0
    fi

    if [ -z "$tg_key" ] || [ -z "$tg_base" ]; then
        return 1
    fi

    local endpoint="${TOKENGATE_STATUSLINE_ENDPOINT:-${tg_base}/v1/statusline}"
    local response
    response=$(fetch_url_json "$endpoint")
    if [ -n "$response" ] && echo "$response" | jq -e '.ok == true' >/dev/null 2>&1; then
        echo "$response" > "$cache_file"
        echo "$response"
        return 0
    fi

    response=$(fetch_url_json "${tg_base}/v1/usage")
    if [ -n "$response" ] && echo "$response" | jq -e '.usage or .quota or .rate_limits' >/dev/null 2>&1; then
        echo "$response" > "$cache_file"
        echo "$response"
        return 0
    fi

    if [ -f "$cache_file" ]; then
        cat "$cache_file"
        return 0
    fi
    return 1
}

tokengate_parts=()
append_statusline_budget() {
    local data="$1"
    local period="$2"
    local label="$3"
    local budget
    budget=$(echo "$data" | jq -r ".budgets.${period} // empty" 2>/dev/null)
    if [ -n "$budget" ] && [ "$budget" != "null" ]; then
        local pct used limit color bar
        pct=$(echo "$budget" | jq -r '.percent // 0' | awk '{printf "%.0f", $1}')
        used=$(echo "$budget" | jq -r '.used // "0"')
        limit=$(echo "$budget" | jq -r '.limit // "0"')
        color=$(usage_color "$pct")
        bar=$(dot_bar "$pct" "$tg_blocks")
        tokengate_parts+=("${white}${label}${reset} ${color}${bar}${reset} ${orange}$(format_money "$used")/$(format_money "$limit") ${pct}%${reset}")
    fi
}

append_usage_rate_limit() {
    local item="$1"
    local window label pct used limit color bar
    window=$(echo "$item" | jq -r '.window // empty')
    used=$(echo "$item" | jq -r '.used // 0')
    limit=$(echo "$item" | jq -r '.limit // 0')
    [ -z "$window" ] || [ "$limit" = "0" ] && return
    case "$window" in
        1d) label="day" ;;
        7d) label="week" ;;
        5h) label="5h" ;;
        *) label="$window" ;;
    esac
    pct=$(awk "BEGIN {printf \"%.0f\", ($used / $limit) * 100}")
    color=$(usage_color "$pct")
    bar=$(dot_bar "$pct" "$tg_blocks")
    tokengate_parts+=("${white}${label}${reset} ${color}${bar}${reset} ${orange}$(format_money "$used")/$(format_money "$limit") ${pct}%${reset}")
}

extract_last_30d_cost() {
    local data="$1"
    echo "$data" | jq -r '
        .cost.last_30_days //
        .cost.last30_days //
        .cost.last_30d //
        .cost.thirty_days //
        .cost["30d"] //
        .usage.last_30_days.cost //
        .usage.last_30_days.actual_cost //
        .usage.last30_days.cost //
        .usage.last30_days.actual_cost //
        .usage.last_30d.cost //
        .usage.last_30d.actual_cost //
        .usage.thirty_days.cost //
        .usage.thirty_days.actual_cost //
        .usage["30d"].cost //
        .usage["30d"].actual_cost //
        empty
    ' 2>/dev/null
}

build_tokengate_parts() {
    local data="$1"
    tokengate_parts=()

    if echo "$data" | jq -e '.ok == true' >/dev/null 2>&1; then
        local today last_30d
        today=$(echo "$data" | jq -r '.cost.today // empty')
        if [ -n "$today" ] && [ "$today" != "null" ]; then
            tokengate_parts+=("${magenta}$(format_money "$today") today${reset}")
        fi
        last_30d=$(extract_last_30d_cost "$data")
        if [ -n "$last_30d" ] && [ "$last_30d" != "null" ]; then
            tokengate_parts+=("${magenta}$(format_money "$last_30d") 30d${reset}")
        fi
        append_statusline_budget "$data" "monthly" "month"
        append_statusline_budget "$data" "weekly" "week"
        append_statusline_budget "$data" "daily" "day"
    else
        local today last_30d
        today=$(echo "$data" | jq -r '.usage.today.cost // .usage.today.actual_cost // empty' 2>/dev/null)
        if [ -n "$today" ] && [ "$today" != "null" ]; then
            tokengate_parts+=("${magenta}$(format_money "$today") today${reset}")
        fi
        last_30d=$(extract_last_30d_cost "$data")
        if [ -n "$last_30d" ] && [ "$last_30d" != "null" ]; then
            tokengate_parts+=("${magenta}$(format_money "$last_30d") 30d${reset}")
        fi

        local quota_limit quota_used
        quota_limit=$(echo "$data" | jq -r '.quota.limit // empty' 2>/dev/null)
        quota_used=$(echo "$data" | jq -r '.quota.used // empty' 2>/dev/null)
        if [ -n "$quota_limit" ] && [ "$quota_limit" != "null" ] && [ "$quota_limit" != "0" ]; then
            local pct color bar
            pct=$(awk "BEGIN {printf \"%.0f\", ($quota_used / $quota_limit) * 100}")
            color=$(usage_color "$pct")
            bar=$(dot_bar "$pct" "$tg_blocks")
            tokengate_parts+=("${white}total${reset} ${color}${bar}${reset} ${orange}$(format_money "$quota_used")/$(format_money "$quota_limit") ${pct}%${reset}")
        fi

        local count i item
        count=$(echo "$data" | jq '.rate_limits | length' 2>/dev/null || echo 0)
        for (( i=0; i<count; i++ )); do
            item=$(echo "$data" | jq ".rate_limits[$i]")
            append_usage_rate_limit "$item"
        done
    fi
}

build_tokengate_output() {
    local level=$1
    local s=" ${dim}|${reset} "
    [ "$level" -gt 1 ] && s="${dim}|${reset}"

    local o=""
    o+="${blue}${model_name}${reset}"
    [ -n "$display_dir" ] && o+="${s}${cyan}${display_dir}${reset}"
    o+="${s}${orange}ctx ${used_tokens}/${total_tokens} ${ctx_pct}%${reset}"

    if [ "${#tokengate_parts[@]}" -gt 0 ]; then
        local part
        for part in "${tokengate_parts[@]}"; do
            o+="${s}${part}"
        done
    else
        o+="${s}${dim}TokenGate unavailable${reset}"
    fi

    echo "$o"
}

render_tokengate_mode() {
    local data out
    data=$(fetch_tokengate_data)
    if [ -n "$data" ]; then
        build_tokengate_parts "$data"
    fi

    out=$(build_tokengate_output 1)
    if [ "$(visible_len "$out")" -gt "$cols" ]; then
        out=$(build_tokengate_output 2)
    fi
    printf "%b" "$out"
}

five_pct="" ; five_reset="" ; five_color="" ; five_bar=""
seven_pct="" ; seven_reset="" ; seven_color="" ; seven_bar=""
extra_part=""

fetch_oauth_usage() {
    local cache_file="$cache_dir/statusline-usage-cache.json"
    local usage_data
    usage_data=$(fetch_cached_json "$cache_file" "$tg_poll")

    if [ -z "$usage_data" ]; then
        local token
        token=$(get_oauth_token)
        if [ -n "$token" ] && [ "$token" != "null" ]; then
            local cc_version response
            cc_version=$(read_json '.version // "2.1.0"')
            response=$(curl -s --max-time 5 \
                -H "Accept: application/json" \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer $token" \
                -H "anthropic-beta: oauth-2025-04-20" \
                -H "User-Agent: claude-code/${cc_version}" \
                "https://api.anthropic.com/api/oauth/usage" 2>/dev/null)
            if [ -n "$response" ] && echo "$response" | jq -e '.five_hour' >/dev/null 2>&1; then
                usage_data="$response"
                echo "$response" > "$cache_file"
            fi
        fi
    fi

    if [ -n "$usage_data" ] && echo "$usage_data" | jq -e '.five_hour' >/dev/null 2>&1; then
        five_pct=$(echo "$usage_data" | jq -r '.five_hour.utilization // 0' | awk '{printf "%.0f", $1}')
        five_reset_iso=$(echo "$usage_data" | jq -r '.five_hour.resets_at // empty')
        five_reset=$(format_reset_time "$five_reset_iso" "time")
        five_color=$(usage_color "$five_pct")
        five_bar=$(dot_bar "$five_pct" "$tg_blocks")

        seven_pct=$(echo "$usage_data" | jq -r '.seven_day.utilization // 0' | awk '{printf "%.0f", $1}')
        seven_reset_iso=$(echo "$usage_data" | jq -r '.seven_day.resets_at // empty')
        seven_reset=$(format_reset_time "$seven_reset_iso" "datetime")
        seven_color=$(usage_color "$seven_pct")
        seven_bar=$(dot_bar "$seven_pct" "$tg_blocks")

        extra_enabled=$(echo "$usage_data" | jq -r '.extra_usage.is_enabled // false')
        if [ "$extra_enabled" = "true" ]; then
            extra_pct=$(echo "$usage_data" | jq -r '.extra_usage.utilization // 0' | awk '{printf "%.0f", $1}')
            extra_used=$(echo "$usage_data" | jq -r '.extra_usage.used_credits // 0' | LC_NUMERIC=C awk '{printf "%.2f", $1/100}')
            extra_limit=$(echo "$usage_data" | jq -r '.extra_usage.monthly_limit // 0' | LC_NUMERIC=C awk '{printf "%.2f", $1/100}')
            extra_color=$(usage_color "$extra_pct")
            extra_part="${white}extra${reset} ${extra_color}\$${extra_used}/\$${extra_limit}${reset}"
        fi
    fi
}

build_claude_output() {
    local level=$1
    local s=" ${dim}|${reset} "
    [ "$level" -gt 1 ] && s="${dim}|${reset}"

    local o=""
    o+="${blue}${model_name}${reset}"
    [ -n "$display_dir" ] && o+="${s}${cyan}${display_dir}${reset}"
    o+="${s}${orange}${used_tokens}/${total_tokens}${reset}"
    o+="${s}${ctx_color}${ctx_pct}% used${reset}"
    o+="${s}${dim}${ctx_remain}% remain${reset}"

    if [ -n "$five_pct" ]; then
        if [ "$level" -le 2 ]; then
            o+="${s}${white}5h${reset} ${five_color}${five_bar}${reset} ${five_color}${five_pct}%${reset}"
        else
            o+="${s}${white}5h${reset} ${five_color}${five_pct}%${reset}"
        fi
        [ -n "$five_reset" ] && o+=" ${dim}@${five_reset}${reset}"
    fi

    if [ -n "$seven_pct" ]; then
        if [ "$level" -le 2 ]; then
            o+="${s}${white}7d${reset} ${seven_color}${seven_bar}${reset} ${seven_color}${seven_pct}%${reset}"
        else
            o+="${s}${white}7d${reset} ${seven_color}${seven_pct}%${reset}"
        fi
        [ -n "$seven_reset" ] && o+=" ${dim}@${seven_reset}${reset}"
    fi

    [ -n "$extra_part" ] && o+="${s}${extra_part}"

    if $session_cost_nonzero && [ "${#tokengate_parts[@]}" -eq 0 ]; then
        cost_disp=$(LC_NUMERIC=C awk "BEGIN {printf \"\$%.4f\", $session_cost}")
        o+="${s}${magenta}${cost_disp} session${reset}"
    fi

    echo "$o"
}

render_claude_mode() {
    local billing_mode="${TOKENGATE_BILLING_MODE:-}"
    if [ -z "$billing_mode" ]; then
        if [ -n "$tg_base" ] && [ -n "$tg_key" ]; then
            billing_mode="AUTO"
        elif [ -n "$ANTHROPIC_API_KEY" ] && [[ "$ANTHROPIC_API_KEY" != tg_* ]]; then
            billing_mode="API_USAGE_DIRECT"
        else
            billing_mode="MONTHLY_SUBSCRIPTION"
        fi
    fi

    if [ "$billing_mode" = "MONTHLY_SUBSCRIPTION" ]; then
        fetch_oauth_usage
    elif [ -n "$tg_key" ] && [ -n "$tg_base" ]; then
        local data
        data=$(fetch_tokengate_data)
        if [ -n "$data" ]; then
            build_tokengate_parts "$data"
        elif [ -n "$tg_key" ]; then
            tokengate_parts+=("${dim}TokenGate: unavailable${reset}")
        fi
    fi

    local out s part
    out=$(build_claude_output 1)
    if [ "${#tokengate_parts[@]}" -gt 0 ]; then
        s=" ${dim}|${reset} "
        for part in "${tokengate_parts[@]}"; do out+="${s}${part}"; done
    fi
    if [ "$(visible_len "$out")" -gt "$cols" ]; then
        out=$(build_claude_output 2)
        if [ "${#tokengate_parts[@]}" -gt 0 ]; then
            s="${dim}|${reset}"
            for part in "${tokengate_parts[@]}"; do out+="${s}${part}"; done
        fi
    fi
    if [ "$(visible_len "$out")" -gt "$cols" ]; then
        out=$(build_claude_output 3)
        if [ "${#tokengate_parts[@]}" -gt 0 ]; then
            s="${dim}|${reset}"
            for part in "${tokengate_parts[@]}"; do out+="${s}${part}"; done
        fi
    fi

    printf "%b" "$out"
}

if [ "$statusline_mode" = "claude" ]; then
    render_claude_mode
else
    render_tokengate_mode
fi

exit 0
