require 'net/http'
require 'uri'
require 'json'

module Kenaito
  module Conf
    class Client
      attr_reader :configs
      attr_accessor :stop

      def initialize(config)
        @config = config
        @cache = Cache.new(config.cache_dir)
        @listener = Listener.new
        @configs = {}
        @last_time = 0
        @started = false
        @stop = false
      end

      def on_change(&block)
        @listener.on_change(&block)
      end

      def get(key)
        @configs[key]
      end

      def get_all
        @configs
      end

      def start
        return if @started
        @started = true

        begin
          fetch_all
        rescue => e
          $stderr.puts "[ConfigClient] fetch all configs error: #{e.message}, trying cache"
          cached = @cache.read
          if cached
            apply_configs(cached)
            $stderr.puts "[ConfigClient] loaded #{@configs.size} configs from cache"
          else
            raise "fetch configs failed and no cache: #{e.message}"
          end
        end
      end

      def stop_client
        @stop = true
      end

      def sync_once
        begin
          fetch_changed
        rescue => e
          $stderr.puts "[ConfigClient] sync error: #{e.message}"
        end
      end

      private

      def http_post(url, payload)
        uri = URI.parse(url)
        http = Net::HTTP.new(uri.host, uri.port)
        http.use_ssl = (uri.scheme == 'https')
        http.open_timeout = 10
        http.read_timeout = 10

        request = Net::HTTP::Post.new(uri.request_uri, 'Content-Type' => 'application/json')
        request.body = payload.to_json

        response = http.request(request)

        if response.code.to_i != 200
          raise "HTTP #{response.code}"
        end

        begin
          JSON.parse(response.body)
        rescue => e
          raise "JSON decode error: #{e.message}"
        end
      end

      def fetch_all
        url = "#{@config.base_url}/api/fetchAllConfigs"
        payload = {
          appName: @config.app_name,
          env: @config.env,
          version: @config.version,
          ignoreGlobal: @config.ignore_global,
          token: @config.token
        }

        data = http_post(url, payload)

        if data['code'] && data['code'] != 0
          raise "server error: #{data['msg'] || 'unknown'}"
        end

        if data['latest.fetch.timestamp']
          @last_time = data['latest.fetch.timestamp'].to_i
        end

        @cache.write(data)
        apply_configs(data)

        $stderr.puts "[ConfigClient] fetched #{@configs.size} configs"
      end

      def fetch_changed
        url = "#{@config.base_url}/api/fetchChangedConfigs"
        payload = {
          appName: @config.app_name,
          env: @config.env,
          version: @config.version,
          lastTime: @last_time,
          ignoreGlobal: @config.ignore_global,
          token: @config.token
        }

        result = http_post(url, payload)

        if result['code'] && result['code'] != 0
          raise "server error code: #{result['code']}"
        end

        changed_data = result['data'] || {}

        if changed_data['latest.fetch.timestamp']
          @last_time = changed_data['latest.fetch.timestamp'].to_i
          changed_data.delete('latest.fetch.timestamp')
        end

        # 处理删除的配置
        if changed_data['deletedKeys'] && !changed_data['deletedKeys'].empty?
          deleted_keys = changed_data['deletedKeys'].split(',')
          deleted_keys.each { |key| @configs.delete(key) }
          $stderr.puts "[ConfigClient] #{deleted_keys.size} configs deleted: #{deleted_keys.join(', ')}"
          changed_data.delete('deletedKeys')
        end

        # 处理变更的配置
        changed = {}
        changed_data.each do |key, value|
          if value.is_a?(String)
            @configs[key] = value
            changed[key] = value
          end
        end

        unless changed.empty?
          @listener.notify(changed)
          $stderr.puts "[ConfigClient] #{changed.size} configs changed: #{changed.keys.join(', ')}"
        end
      end

      def apply_configs(data)
        data.each do |key, value|
          next if key.start_with?('kenaito.') || key == 'latest.fetch.timestamp'
          if value.is_a?(String)
            @configs[key] = value
          end
        end
      end
    end
  end
end
