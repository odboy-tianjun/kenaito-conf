# Kenaito Config Center Ruby Client

[English](README_EN.md) | 中文

Kenaito 配置中心 Ruby 客户端 SDK，适用于 Ruby 2.5+ 项目。

## 安装

### 方式一：本地安装

```bash
cd kenaito-conf-ruby-client
gem build kenaito-conf.gemspec
gem install kenaito-conf-1.0.3.gem
```

### 方式二：Bundler

在你的项目 `Gemfile` 中添加：

```ruby
gem 'kenaito-conf', path: '/path/to/kenaito-conf-ruby-client'
```

然后执行：

```bash
bundle install
```

### 方式三：直接引入

```ruby
require_relative '/path/to/kenaito-conf-ruby-client/lib/kenaito_conf'
```

## 使用示例

```ruby
require 'kenaito_conf'

config = Kenaito::Conf::Config.new(
  base_url: 'http://localhost:8080',
  app_name: 'my-app',
  env: 'daily',
  sync_interval: 10
)

client = Kenaito::Conf::Client.new(config)

# 监听配置变更
client.on_change do |changed|
  changed.each do |key, value|
    puts "config changed: #{key} = #{value}"
  end
end

# 启动客户端（拉取配置）
client.start

# 获取配置
value = client.get('database.host')
all_configs = client.get_all

# 手动同步一次
client.sync_once

# 停止
client.stop_client
```

## 配置项

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| base_url | 配置中心后端地址 | http://localhost:8080 |
| app_name | 应用名称 | 空 |
| env | 环境（dev/daily/prod） | dev |
| version | 配置版本 | latest |
| token | 应用 Token | 空 |
| sync_interval | 同步间隔（秒） | 10 |
| ignore_global | 是否忽略全局配置 | false |
| cache_dir | 缓存目录 | . |

## 功能特性

- ✅ 全量拉取配置
- ✅ 增量同步变更
- ✅ 本地文件缓存容灾
- ✅ 配置变更监听回调
- ✅ 删除配置感知（deletedKeys）
