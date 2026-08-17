# Kenaito Config Center Ruby Client

[English](README_EN.md) | 中文

Kenaito Config Center Ruby client SDK, compatible with Ruby 2.5+.

## Installation

### Option 1: Local Install

```bash
cd kenaito-conf-ruby-client
gem build kenaito-conf.gemspec
gem install kenaito-conf-1.0.3.gem
```

### Option 2: Bundler

Add to your project `Gemfile`:

```ruby
gem 'kenaito-conf', path: '/path/to/kenaito-conf-ruby-client'
```

Then run:

```bash
bundle install
```

### Option 3: Direct Require

```ruby
require_relative '/path/to/kenaito-conf-ruby-client/lib/kenaito_conf'
```

## Usage Example

```ruby
require 'kenaito_conf'

config = Kenaito::Conf::Config.new(
  base_url: 'http://localhost:8080',
  app_name: 'my-app',
  env: 'daily',
  sync_interval: 10
)

client = Kenaito::Conf::Client.new(config)

# Listen for config changes
client.on_change do |changed|
  changed.each do |key, value|
    puts "config changed: #{key} = #{value}"
  end
end

# Start client (fetch configs)
client.start

# Get config
value = client.get('database.host')
all_configs = client.get_all

# Manual sync once
client.sync_once

# Stop
client.stop_client
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| base_url | Config center backend URL | http://localhost:8080 |
| app_name | Application name | empty |
| env | Environment (dev/daily/prod) | dev |
| version | Config version | latest |
| token | Application token | empty |
| sync_interval | Sync interval (seconds) | 10 |
| ignore_global | Ignore global configs | false |
| cache_dir | Cache directory | . |

## Features

- ✅ Full config fetch
- ✅ Incremental sync
- ✅ Local file cache for disaster recovery
- ✅ Config change listener callback
- ✅ Deleted config awareness (deletedKeys)
