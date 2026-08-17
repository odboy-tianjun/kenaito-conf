require_relative '../kenaito-conf-ruby-client/lib/kenaito_conf'

def get_env_var(key, default)
  value = ENV[key]
  (value && !value.empty?) ? value : default
end

# 配置值（可被远程配置覆盖）
config_values = {
  'database.host' => get_env_var('DATABASE_HOST', '192.168.1.1'),
  'database.port' => get_env_var('DATABASE_PORT', '8888'),
  'database.name' => get_env_var('DATABASE_NAME', 'xxxx'),
  'database.user' => get_env_var('DATABASE_USER', 'xxxx'),
  'database.password' => get_env_var('DATABASE_PASSWORD', 'xxxxxxxx')
}

config = Kenaito::Conf::Config.new(
  base_url: get_env_var('KENAITO_BASE_URL', 'http://localhost:8080'),
  app_name: get_env_var('KENAITO_APP_NAME', 'kenaito-ruby-demo'),
  env: get_env_var('KENAITO_ENV', 'daily'),
  version: get_env_var('KENAITO_VERSION', 'latest'),
  token: get_env_var('KENAITO_TOKEN', ''),
  sync_interval: get_env_var('KENAITO_SYNC_INTERVAL', '10').to_i,
  cache_dir: get_env_var('KENAITO_CACHE_DIR', '.')
)

client = Kenaito::Conf::Client.new(config)

# 监听配置变更
client.on_change do |changed|
  puts "Config changed:"
  changed.each do |key, value|
    puts "  #{key} = #{value}"
    if config_values.key?(key)
      config_values[key] = value
      puts "  [updated] #{key} = #{value}"
    end
  end
end

# 启动客户端
begin
  client.start
  remote = client.get_all
  config_values.each do |key, _|
    if remote.key?(key)
      config_values[key] = remote[key]
    end
  end
  puts "loaded #{remote.size} configs from remote\n\n"
rescue => e
  puts "Failed to init config client, using local defaults: #{e.message}\n\n"
end

# 显示当前配置
puts "Current configs:"
config_values.keys.sort.each do |key|
  puts "  #{key} = #{config_values[key]}"
end

puts "\nListening for config changes... (press Ctrl+C to stop)\n"

# 信号处理
trap('INT') do
  puts "\nShutting down..."
  client.stop_client
  exit(0)
end

# 同步循环
until client.stop
  client.sync_once
  sleep(config.sync_interval)
end
