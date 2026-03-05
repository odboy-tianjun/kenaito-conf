# frozen_string_literal: true

require_relative '../kenaito-conf-ruby-client/lib/kenaito_conf'

config_values = {
  'database.host' => ENV['DATABASE_HOST'] || '192.168.1.1',
  'database.port' => ENV['DATABASE_PORT'] || '8888',
  'database.name' => ENV['DATABASE_NAME'] || 'xxxx',
  'database.user' => ENV['DATABASE_USER'] || 'xxxx',
  'database.password' => ENV['DATABASE_PASSWORD'] || 'xxxxxxxx'
}

config = KenaitoConf::Config.new(
  base_url: ENV['KENAITO_BASE_URL'] || 'http://localhost:8080',
  app_name: ENV['KENAITO_APP_NAME'] || 'kenaito-springboot-demo',
  env: ENV['KENAITO_PROFILE'] || 'daily',
  version: ENV['KENAITO_VERSION'] || 'latest',
  token: ENV['KENAITO_TOKEN'] || '',
  sync_interval: (ENV['KENAITO_SYNC_INTERVAL'] || '30').to_i,
  cache_dir: ENV['KENAITO_CACHE_DIR'] || '.'
)

client = KenaitoConf::ConfigClient.new(config)

client.on_change do |changed|
  changed.each do |k, v|
    if config_values.key?(k)
      config_values[k] = v
      warn "config updated: #{k} = #{v}"
    end
  end
end

begin
  client.start
  remote = client.get_all
  config_values.each_key do |k|
    config_values[k] = remote[k] if remote[k]
  end
  warn "loaded #{remote.size} configs from remote"
rescue StandardError => e
  warn "failed to init config client: #{e.message}"
end

puts 'Current configs:'
config_values.each do |k, v|
  puts "  #{k} = #{v}"
end

puts "\nListening for config changes... (press Ctrl+C to stop)"
sleep
