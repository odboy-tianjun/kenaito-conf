require 'json'
require 'fileutils'

module Kenaito
  module Conf
    class Cache
      def initialize(cache_dir)
        @cache_file = File.join(cache_dir, 'config-cache.json')
      end

      def read
        return nil unless File.exist?(@cache_file)
        
        begin
          content = File.read(@cache_file)
          JSON.parse(content)
        rescue => e
          $stderr.puts "[ConfigClient] cache read error: #{e.message}"
          nil
        end
      end

      def write(data)
        begin
          FileUtils.mkdir_p(File.dirname(@cache_file))
          File.write(@cache_file, JSON.pretty_generate(data))
        rescue => e
          $stderr.puts "[ConfigClient] cache write error: #{e.message}"
        end
      end
    end
  end
end
