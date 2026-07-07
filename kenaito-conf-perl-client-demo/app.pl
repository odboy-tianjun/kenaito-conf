#!/usr/bin/env perl
use strict;
use warnings;
use FindBin;
use lib "$FindBin::Bin/../../kenaito-conf-perl-client/lib";
use JSON::PP;
use Kenaito::Conf::Config;
use Kenaito::Conf::Client;

my %config_values = (
    'database.host'     => $ENV{DATABASE_HOST}     // '192.168.1.1',
    'database.port'     => $ENV{DATABASE_PORT}     // '8888',
    'database.name'     => $ENV{DATABASE_NAME}     // 'xxxx',
    'database.user'     => $ENV{DATABASE_USER}     // 'xxxx',
    'database.password' => $ENV{DATABASE_PASSWORD} // 'xxxxxxxx',
);

sub get_env {
    my ($key, $default) = @_;
    return defined $ENV{$key} && $ENV{$key} ne '' ? $ENV{$key} : $default;
}

my $cfg = Kenaito::Conf::Config->new(
    base_url      => get_env('KENAITO_BASE_URL',      'http://localhost:8080'),
    app_name      => get_env('KENAITO_APP_NAME',      'kenaito-springboot-demo'),
    env           => get_env('KENAITO_ENV',            'daily'),
    version       => get_env('KENAITO_VERSION',        'latest'),
    token         => get_env('KENAITO_TOKEN',          ''),
    sync_interval => int(get_env('KENAITO_SYNC_INTERVAL', '10')),
    cache_dir     => get_env('KENAITO_CACHE_DIR',      '.'),
);

my $client = Kenaito::Conf::Client->new(config => $cfg);

$client->on_change(sub {
    my ($changed) = @_;
    print "Config changed:\n";
    for my $k (sort keys %$changed) {
        my $v = $changed->{$k};
        print "  $k = $v\n";
        if (exists $config_values{$k}) {
            $config_values{$k} = $v;
            print "  [updated] $k = $v\n";
        }
    }
});

eval {
    $client->start();
    1;
} or do {
    warn "Failed to init config client, using local defaults: $@\n";
};

my $remote = $client->get_all();
for my $k (keys %config_values) {
    if (exists $remote->{$k}) {
        $config_values{$k} = $remote->{$k};
    }
}

my $count = scalar keys %$remote;
print "loaded $count configs from remote\n\n";

print "Current configs:\n";
for my $k (sort keys %config_values) {
    print "  $k = $config_values{$k}\n";
}

print "\nListening for config changes... (press Ctrl+C to stop)\n";

$SIG{INT} = sub {
    print "\nShutting down...\n";
    $client->stop();
    exit 0;
};

# sync loop
while (1) {
    sleep $cfg->{sync_interval};
    last if $client->{stop};
    $client->sync_once();
}
