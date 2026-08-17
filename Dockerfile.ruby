FROM ruby:3.3-slim

WORKDIR /kenaito-conf-ruby-client-demo

# Copy client library
COPY kenaito-conf-ruby-client/ /kenaito-conf-ruby-client/

# Copy demo
COPY kenaito-conf-ruby-client-demo/app.rb /kenaito-conf-ruby-client-demo/app.rb

ENTRYPOINT ["ruby", "app.rb"]
