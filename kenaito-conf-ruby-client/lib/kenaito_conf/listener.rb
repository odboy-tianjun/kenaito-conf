module Kenaito
  module Conf
    class Listener
      def initialize
        @handlers = []
      end

      def on_change(&block)
        @handlers << block
      end

      def notify(changed)
        @handlers.each do |handler|
          begin
            handler.call(changed)
          rescue => e
            $stderr.puts "[ConfigClient] change handler error: #{e.message}"
          end
        end
      end
    end
  end
end
