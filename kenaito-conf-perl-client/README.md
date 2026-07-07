# Kenaito Config Center Perl Client

Perl 客户端 SDK，适用于 Perl 5.16+ 项目。

## 安装

### 从本地安装

```bash
cd kenaito-conf-perl-client
cpanm .
```

### 开发模式（符号链接）

在 demo 项目中添加 lib 路径：

```perl
use lib '/path/to/kenaito-conf-perl-client/lib';
```

## 使用

```perl
use Kenaito::Conf::Config;
use Kenaito::Conf::Client;

my $cfg = Kenaito::Conf::Config->new(
    base_url      => 'http://localhost:8080',
    app_name      => 'my-app',
    env           => 'daily',
    version       => 'latest',
    token         => '',
    sync_interval => 10,
    cache_dir     => '.',
);

my $client = Kenaito::Conf::Client->new(config => $cfg);

# 监听配置变更
$client->on_change(sub {
    my ($changed) = @_;
    for my $k (keys %$changed) {
        print "config changed: $k = $changed->{$k}\n";
    }
});

# 启动客户端（拉取配置）
$client->start();

# 获取配置
my $value = $client->get('database.host');
my $all   = $client->get_all();

# 手动同步一次
$client->sync_once();

# 停止
$client->stop();
```

## 配置项

| 参数            | 说明         | 默认值                     |
|----------------|------------|-------------------------|
| `base_url`     | 配置中心地址     | `http://localhost:8080` |
| `app_name`     | 应用名称       | （必填）                   |
| `env`          | 环境标识       | `dev`                   |
| `version`      | 配置版本       | `latest`                |
| `token`        | 应用 Token   | 空                       |
| `sync_interval`| 同步间隔（秒）    | `60`                    |
| `ignore_global`| 忽略全局配置     | `0`                     |
| `cache_dir`    | 缓存目录       | `.`                     |

## 依赖

- Perl 5.16+
- HTTP::Tiny
- JSON::PP

## Demo

```bash
cd kenaito-conf-perl-client-demo
KENAITO_BASE_URL=http://localhost:8080 \
KENAITO_APP_NAME=kenaito-springboot-demo \
KENAITO_ENV=daily \
perl app.pl
```
