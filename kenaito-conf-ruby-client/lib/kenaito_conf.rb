# frozen_string_literal: true

require 'net/http'
require 'json'

module KenaitoConf
  class Config
    attr_accessor :base_url, :app_name, :env, :version, :token, :sync_interval, :ignore_global, :cache_dir

    def initialize(options = {})
      @base_url = options[:base_url] || ''
      @app_name = options[:app_name] || ''
      @env = options[:env] || 'dev'
      @version = options[:version] || 'latest'
      @token = options[:token] || ''
      @sync_interval = options[:sync_interval] || 60
      @ignore_global = options[:ignore_global] || false
      @cache_dir = options[:cache_dir] || '.'
    end
  end

  class Cache
    CACHE_FILE = 'config-cache.json'

    def initialize(cache_dir = '.')
      @cache_dir = cache_dir
    end

    def write(data)
      File.write(File.join(@cache_dir, CACHE_FILE), JSON.pretty_generate(data))
    end

    def read
      path = File.join(@cache_dir, CACHE_FILE)
      return nil unless File.exist?(path)
      JSON.parse(File.read(path))
    rescue StandardError
      nil
    end
  end

  class ConfigClient
    def initialize(config)
      @config = config
      @cache = Cache.new(config.cache_dir)
      @configs = {}
      @last_time = 0
      @handlers = []
      @started = false
    end

    def on_change(&handler)
      @handlers << handler
    end

    def get(key)
      @configs[key]
    end

    def get_all
      @configs.dup
    end

    def start
      return if @started
      @started = true

      begin
        fetch_all
      rescue StandardError => e
        warn "[ConfigClient] fetch all configs error: #{e.message}, trying cache"
        cached = @cache.read
        if cached
          apply_configs(cached)
          warn "[ConfigClient] loaded #{@configs.size} configs from cache"
        else
          raise "fetch configs failed and no cache: #{e.message}"
        end
      end

      Thread.new do
        loop do
          sleep @config.sync_interval
          begin
            fetch_changed
          rescue StandardError => e
            warn "[ConfigClient] sync error: #{e.message}"
          end
        end
      end
    end

    def stop
      @started = false
    end

    private

    def fetch_all
      uri = URI("#{@config.base_url}/api/fetchAllConfigs")
      http = Net::HTTP.new(uri.host, uri.port)
      http.read_timeout = 10

      request = Net::HTTP::Post.new(uri.path, 'Content-Type' => 'application/json')
      request.body = JSON.generate({
        appName: @config.app_name,
        env: @config.env,
        version: @config.version,
        ignoreGlobal: @config.ignore_global,
        token: @config.token
      })

      response = http.request(request)
      raise "HTTP #{response.code}" unless response.code == '200'

      data = JSON.parse(response.body)
      raise "server error: #{data['msg']}" if data['code'] && data['code'] != 0

      @last_time = data['latest.fetch.timestamp'].to_i if data['latest.fetch.timestamp']

      @cache.write(data)
      apply_configs(data)
      warn "[ConfigClient] fetched #{@configs.size} configs"
    end

    def fetch_changed
      uri = URI("#{@config.base_url}/api/fetchChangedConfigs")
      http = Net::HTTP.new(uri.host, uri.port)
      http.read_timeout = 10

      request = Net::HTTP::Post.new(uri.path, 'Content-Type' => 'application/json')
      request.body = JSON.generate({
        appName: @config.app_name,
        env: @config.env,
        version: @config.version,
        lastTime: @last_time,
        ignoreGlobal: @config.ignore_global,
        token: @config.token
      })

      response = http.request(request)
      result = JSON.parse(response.body)
      raise "server error" if !result['code'] || result['code'] != 0

      changed_data = result['data'] || {}
      @last_time = changed_data.delete('latest.fetch.timestamp').to_i if changed_data['latest.fetch.timestamp']

      changed = {}
      changed_data.each do |k, v|
        changed[k] = v
      end

      @configs.merge!(changed)
      @handlers.each { |h| h.call(changed) rescue warn "[ConfigClient] handler error: #{$!.message}" }
    end

    def apply_configs(data)
      data.each do |k, v|
        next if k.start_with?('kenaito.') || k == 'latest.fetch.timestamp'
        next unless v.is_a?(String)
        @configs[k] = v
      end
    end
  end
end
