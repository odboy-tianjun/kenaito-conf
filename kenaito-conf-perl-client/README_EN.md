# Kenaito Config Center Perl Client

Perl client SDK for the Kenaito Config Center, suitable for Perl 5.16+ projects.

## Installation

### Install from Local

```bash
cd kenaito-conf-perl-client
cpanm .
```

### Development Mode (Symlink)

Add the lib path in your demo project:

```perl
use lib '/path/to/kenaito-conf-perl-client/lib';
```

## Usage

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

# Listen for config changes
$client->on_change(sub {
    my ($changed) = @_;
    for my $k (keys %$changed) {
        print "config changed: $k = $changed->{$k}\n";
    }
});

# Start client (fetch configs)
$client->start();

# Get config
my $value = $client->get('database.host');
my $all   = $client->get_all();

# Manual sync
$client->sync_once();

# Stop
$client->stop();
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `base_url` | Config center URL | `http://localhost:8080` |
| `app_name` | Application name | (Required) |
| `env` | Environment | `dev` |
| `version` | Config version | `latest` |
| `token` | App token | Empty |
| `sync_interval` | Sync interval (seconds) | `60` |
| `ignore_global` | Whether to ignore global configs | `0` |
| `cache_dir` | Cache directory | `.` |

## Dependencies

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
