module Kenaito
  module Conf
    class Config
      attr_accessor :base_url, :app_name, :env, :version, :token, :sync_interval, :ignore_global, :cache_dir

      def initialize(options = {})
        @base_url = options[:base_url] || 'http://localhost:8080'
        @app_name = options[:app_name] || ''
        @env = options[:env] || 'dev'
        @version = options[:version] || 'latest'
        @token = options[:token] || ''
        @sync_interval = options[:sync_interval] || 10
        @ignore_global = options[:ignore_global] || false
        @cache_dir = options[:cache_dir] || '.'
      end
    end
  end
end
