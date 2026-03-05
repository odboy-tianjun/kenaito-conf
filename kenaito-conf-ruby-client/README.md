# Kenaito Config Center Ruby Client

Ruby 语言的配置中心客户端 SDK。

## 依赖

- Ruby 2.7+
- 标准库（net/http, json, openssl, base64）

## 安装

```bash
gem build kenaito-conf.gemspec
gem install kenaito-conf-1.0.0.gem
```

或在 Gemfile 中添加：

```ruby
gem 'kenaito-conf', path: '../kenaito-conf-ruby-client'
```

## 使用

```ruby
require 'kenaito_conf'

config = KenaitoConf::Config.new(
  base_url: 'http://localhost:8080',
  app_name: 'kenaito-springboot-demo',
  env: 'daily',
  version: 'latest',
  sync_interval: 30
)

client = KenaitoConf::ConfigClient.new(config)

client.on_change do |changed|
  changed.each do |k, v|
    puts "config changed: #{k} = #{v}"
  end
end

client.start

value = client.get('database.host')
all = client.get_all

client.stop
```

## 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `base_url` | 配置中心地址 | 必填 |
| `app_name` | 应用名称 | 必填 |
| `env` | 环境 | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 认证令牌 | 空 |
| `sync_interval` | 同步间隔（秒） | `60` |
| `ignore_global` | 是否忽略全局配置 | `false` |
| `cache_dir` | 本地缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-ruby-client-demo
ruby app.rb
```
