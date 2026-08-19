#!/usr/bin/env bash
# Kenaito Config Center Bash Client
# 依赖: curl, jq
# 用法: source kenaito_conf.sh 后调用函数

KENAITO_CONF_VERSION="1.0.3"

# 配置变量（使用前需设置）
KENAITO_BASE_URL="${KENAITO_BASE_URL:-http://localhost:8080}"
KENAITO_APP_NAME="${KENAITO_APP_NAME:-}"
KENAITO_ENV="${KENAITO_ENV:-dev}"
KENAITO_VERSION="${KENAITO_VERSION:-latest}"
KENAITO_TOKEN="${KENAITO_TOKEN:-}"
KENAITO_SYNC_INTERVAL="${KENAITO_SYNC_INTERVAL:-10}"
KENAITO_IGNORE_GLOBAL="${KENAITO_IGNORE_GLOBAL:-false}"
KENAITO_CACHE_DIR="${KENAITO_CACHE_DIR:-.}"

# 内部状态
_KENAITO_LAST_TIME=0
_KENAITO_STARTED=false
declare -A _KENAITO_CONFIGS

# 获取缓存文件路径
_kenaito_cache_file() {
    echo "${KENAITO_CACHE_DIR}/config-cache.json"
}

# 从缓存读取
_kenaito_cache_read() {
    local cache_file
    cache_file=$(_kenaito_cache_file)
    if [ -f "$cache_file" ]; then
        cat "$cache_file"
        return 0
    fi
    return 1
}

# 写入缓存
_kenaito_cache_write() {
    local data="$1"
    local cache_file
    cache_file=$(_kenaito_cache_file)
    mkdir -p "$(dirname "$cache_file")"
    echo "$data" > "$cache_file"
}

# HTTP POST 请求
_kenaito_http_post() {
    local url="$1"
    local payload="$2"
    local response_file
    local http_code_file
    local http_code

    response_file=$(mktemp)
    http_code_file=$(mktemp)

    curl -s -o "$response_file" -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        --connect-timeout 10 \
        --max-time 10 \
        "$url" > "$http_code_file" 2>/dev/null

    local curl_exit=$?
    http_code=$(cat "$http_code_file")

    if [ $curl_exit -ne 0 ]; then
        echo "ERROR: curl failed with exit code $curl_exit" >&2
        rm -f "$response_file" "$http_code_file"
        return 1
    fi

    if [ "$http_code" != "200" ]; then
        echo "ERROR: HTTP $http_code" >&2
        rm -f "$response_file" "$http_code_file"
        return 1
    fi

    cat "$response_file"
    rm -f "$response_file" "$http_code_file"
}

# 应用配置到本地存储
_kenaito_apply_configs() {
    local data="$1"
    _KENAITO_CONFIGS=()

    local keys
    keys=$(echo "$data" | jq -r 'keys[]' 2>/dev/null)

    for key in $keys; do
        if [[ "$key" == kenaito.* ]] || [[ "$key" == "latest.fetch.timestamp" ]]; then
            continue
        fi
        local value
        value=$(echo "$data" | jq -r --arg k "$key" '.[$k] // empty' 2>/dev/null)
        if [ -n "$value" ]; then
            _KENAITO_CONFIGS["$key"]="$value"
        fi
    done
}

# 获取配置值
kenaito_get() {
    local key="$1"
    echo "${_KENAITO_CONFIGS[$key]:-}"
}

# 获取所有配置（输出为 key=value 格式）
kenaito_get_all() {
    for key in "${!_KENAITO_CONFIGS[@]}"; do
        echo "$key=${_KENAITO_CONFIGS[$key]}"
    done
}

# 全量拉取配置
kenaito_fetch_all() {
    local url="${KENAITO_BASE_URL}/api/fetchAllConfigs"
    local payload
    payload=$(jq -c -n \
        --arg appName "$KENAITO_APP_NAME" \
        --arg env "$KENAITO_ENV" \
        --arg version "$KENAITO_VERSION" \
        --argjson ignoreGlobal "$KENAITO_IGNORE_GLOBAL" \
        --arg token "$KENAITO_TOKEN" \
        '{appName: $appName, env: $env, version: $version, ignoreGlobal: $ignoreGlobal, token: $token}')

    local data
    data=$(_kenaito_http_post "$url" "$payload")
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        return 1
    fi

    local code
    code=$(echo "$data" | jq -r '.code // empty' 2>/dev/null)
    if [ -n "$code" ] && [ "$code" != "0" ]; then
        local msg
        msg=$(echo "$data" | jq -r '.msg // "unknown"' 2>/dev/null)
        echo "ERROR: server error: $msg" >&2
        return 1
    fi

    local timestamp
    timestamp=$(echo "$data" | jq -r '."latest.fetch.timestamp" // empty' 2>/dev/null)
    if [ -n "$timestamp" ]; then
        _KENAITO_LAST_TIME=$timestamp
    fi

    _kenaito_cache_write "$data"
    _kenaito_apply_configs "$data"

    local count=${#_KENAITO_CONFIGS[@]}
    echo "[ConfigClient] fetched $count configs" >&2
}

# 增量拉取变更
kenaito_fetch_changed() {
    local url="${KENAITO_BASE_URL}/api/fetchChangedConfigs"
    local payload
    payload=$(jq -c -n \
        --arg appName "$KENAITO_APP_NAME" \
        --arg env "$KENAITO_ENV" \
        --arg version "$KENAITO_VERSION" \
        --argjson lastTime "$_KENAITO_LAST_TIME" \
        --argjson ignoreGlobal "$KENAITO_IGNORE_GLOBAL" \
        --arg token "$KENAITO_TOKEN" \
        '{appName: $appName, env: $env, version: $version, lastTime: $lastTime, ignoreGlobal: $ignoreGlobal, token: $token}')

    local result
    result=$(_kenaito_http_post "$url" "$payload")
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        return 1
    fi

    local code
    code=$(echo "$result" | jq -r '.code // empty' 2>/dev/null)
    if [ -n "$code" ] && [ "$code" != "0" ]; then
        echo "ERROR: server error code: $code" >&2
        return 1
    fi

    local changed_data
    changed_data=$(echo "$result" | jq -c '.data // {}' 2>/dev/null)

    local timestamp
    timestamp=$(echo "$changed_data" | jq -r '."latest.fetch.timestamp" // empty' 2>/dev/null)
    if [ -n "$timestamp" ]; then
        _KENAITO_LAST_TIME=$timestamp
    fi

    # 处理删除的配置
    local deleted_keys
    deleted_keys=$(echo "$changed_data" | jq -r '.deletedKeys // empty' 2>/dev/null)
    if [ -n "$deleted_keys" ]; then
        IFS=',' read -ra keys <<< "$deleted_keys"
        for key in "${keys[@]}"; do
            unset "_KENAITO_CONFIGS[$key]"
        done
        local count=${#keys[@]}
        echo "[ConfigClient] $count configs deleted: $deleted_keys" >&2
    fi

    # 处理变更的配置
    local changed_count=0
    local changed_keys_list=""
    local keys
    keys=$(echo "$changed_data" | jq -r 'keys[]' 2>/dev/null)

    for key in $keys; do
        if [[ "$key" == "latest.fetch.timestamp" ]] || [[ "$key" == "deletedKeys" ]]; then
            continue
        fi
        local value
        value=$(echo "$changed_data" | jq -r --arg k "$key" '.[$k] // empty' 2>/dev/null)
        if [ -n "$value" ]; then
            _KENAITO_CONFIGS["$key"]="$value"
            if [ -n "$changed_keys_list" ]; then
                changed_keys_list="$changed_keys_list, $key"
            else
                changed_keys_list="$key"
            fi
            changed_count=$((changed_count + 1))
        fi
    done

    if [ $changed_count -gt 0 ]; then
        if type kenaito_on_change &>/dev/null; then
            kenaito_on_change "$changed_keys_list"
        fi
        echo "[ConfigClient] $changed_count configs changed: $changed_keys_list" >&2
    fi
}

# 启动客户端
kenaito_start() {
    if [ "$_KENAITO_STARTED" = true ]; then
        return
    fi
    _KENAITO_STARTED=true

    if ! kenaito_fetch_all; then
        echo "[ConfigClient] fetch all configs error, trying cache" >&2
        local cached
        if cached=$(_kenaito_cache_read); then
            _kenaito_apply_configs "$cached"
            local count=${#_KENAITO_CONFIGS[@]}
            echo "[ConfigClient] loaded $count configs from cache" >&2
        else
            echo "[ConfigClient] fetch configs failed and no cache" >&2
            return 1
        fi
    fi
}

# 同步一次
kenaito_sync_once() {
    if ! kenaito_fetch_changed; then
        echo "[ConfigClient] sync error" >&2
    fi
}
