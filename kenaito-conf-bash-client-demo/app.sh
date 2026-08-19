#!/usr/bin/env bash
# Kenaito Config Center Bash Client Demo

# 引入客户端库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../kenaito-conf-bash-client/lib/kenaito_conf.sh"

# 配置（可通过环境变量覆盖）
export KENAITO_BASE_URL="${KENAITO_BASE_URL:-http://localhost:8080}"
export KENAITO_APP_NAME="${KENAITO_APP_NAME:-kenaito-bash-demo}"
export KENAITO_ENV="${KENAITO_ENV:-daily}"
export KENAITO_VERSION="${KENAITO_VERSION:-latest}"
export KENAITO_TOKEN="${KENAITO_TOKEN:-}"
export KENAITO_SYNC_INTERVAL="${KENAITO_SYNC_INTERVAL:-10}"
export KENAITO_CACHE_DIR="${KENAITO_CACHE_DIR:-.}"

# 本地默认配置
declare -A config_values
config_values=(
    ["database.host"]="${DATABASE_HOST:-192.168.1.1}"
    ["database.port"]="${DATABASE_PORT:-8888}"
    ["database.name"]="${DATABASE_NAME:-xxxx}"
    ["database.user"]="${DATABASE_USER:-xxxx}"
    ["database.password"]="${DATABASE_PASSWORD:-xxxxxxxx}"
)

# 变更回调函数
kenaito_on_change() {
    local changed_keys="$1"
    echo "Config changed:"
    IFS=', ' read -ra keys <<< "$changed_keys"
    for key in "${keys[@]}"; do
        local value
        value=$(kenaito_get "$key")
        echo "  $key = $value"
        if [ -n "${config_values[$key]+x}" ]; then
            config_values["$key"]="$value"
            echo "  [updated] $key = $value"
        fi
    done
}

# 启动客户端
kenaito_start
if [ $? -eq 0 ]; then
    # 用远程配置覆盖本地默认值
    for key in "${!config_values[@]}"; do
        remote_value=$(kenaito_get "$key")
        if [ -n "$remote_value" ]; then
            config_values["$key"]="$remote_value"
        fi
    done

    # 统计配置数量
    count=0
    for key in "${!_KENAITO_CONFIGS[@]}"; do
        count=$((count + 1))
    done
    echo "loaded $count configs from remote"
    echo ""
else
    echo "Failed to init config client, using local defaults"
    echo ""
fi

# 显示当前配置
echo "Current configs:"
for key in $(echo "${!config_values[@]}" | tr ' ' '\n' | sort); do
    echo "  $key = ${config_values[$key]}"
done

echo ""
echo "Listening for config changes... (press Ctrl+C to stop)"

# 信号处理
trap 'echo ""; echo "Shutting down..."; exit 0' INT TERM

# 同步循环
while true; do
    kenaito_sync_once
    sleep "$KENAITO_SYNC_INTERVAL"
done
