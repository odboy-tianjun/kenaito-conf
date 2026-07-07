package Kenaito::Conf::Client;
use strict;
use warnings;
use HTTP::Tiny;
use JSON::PP;
use Kenaito::Conf::Config;
use Kenaito::Conf::Cache;
use Kenaito::Conf::Listener;

sub new {
    my ($class, %args) = @_;
    my $config = $args{config};
    die "config is required" unless $config;

    my $self = bless {
        config   => $config,
        cache    => Kenaito::Conf::Cache->new($config->{cache_dir}),
        listener => Kenaito::Conf::Listener->new(),
        configs  => {},
        last_time => 0,
        started  => 0,
        stop     => 0,
        http     => HTTP::Tiny->new(timeout => 10),
    }, $class;

    return $self;
}

sub on_change {
    my ($self, $handler) = @_;
    $self->{listener}->on_change($handler);
}

sub get {
    my ($self, $key) = @_;
    return $self->{configs}{$key};
}

sub get_all {
    my ($self) = @_;
    return { %{$self->{configs}} };
}

sub start {
    my ($self) = @_;
    return if $self->{started};
    $self->{started} = 1;

    eval {
        $self->_fetch_all();
        1;
    } or do {
        my $err = $@;
        warn "[ConfigClient] fetch all configs error: $err, trying cache\n";
        my $cached = $self->{cache}->read();
        if ($cached) {
            $self->_apply_configs($cached);
            my $count = scalar keys %{$self->{configs}};
            print "[ConfigClient] loaded $count configs from cache\n";
        } else {
            die "fetch configs failed and no cache: $err";
        }
    };
}

sub stop {
    my ($self) = @_;
    $self->{stop} = 1;
}

sub sync_once {
    my ($self) = @_;
    eval {
        $self->_fetch_changed();
        1;
    } or do {
        warn "[ConfigClient] sync error: $@\n";
    };
}

sub _fetch_all {
    my ($self) = @_;
    my $url = "$self->{config}{base_url}/api/fetchAllConfigs";
    my $payload = JSON::PP->new->utf8->encode({
        appName      => $self->{config}{app_name},
        env          => $self->{config}{env},
        version      => $self->{config}{version},
        ignoreGlobal => $self->{config}{ignore_global} ? \1 : \0,
        token        => $self->{config}{token},
    });

    my $resp = $self->{http}->post($url, {
        content => $payload,
        headers => { 'Content-Type' => 'application/json' },
    });

    die "HTTP error: $resp->{status} $resp->{reason}" unless $resp->{success};

    my $data = JSON::PP->new->utf8->decode($resp->{content});

    if (exists $data->{code} && $data->{code} != 0) {
        die "server error: " . ($data->{msg} // 'unknown');
    }

    if (my $ts = $data->{'latest.fetch.timestamp'}) {
        $self->{last_time} = int($ts);
    }

    $self->{cache}->write($data);
    $self->_apply_configs($data);

    my $count = scalar keys %{$self->{configs}};
    print "[ConfigClient] fetched $count configs\n";
}

sub _fetch_changed {
    my ($self) = @_;
    my $url = "$self->{config}{base_url}/api/fetchChangedConfigs";
    my $payload = JSON::PP->new->utf8->encode({
        appName      => $self->{config}{app_name},
        env          => $self->{config}{env},
        version      => $self->{config}{version},
        lastTime     => $self->{last_time},
        ignoreGlobal => $self->{config}{ignore_global} ? \1 : \0,
        token        => $self->{config}{token},
    });

    my $resp = $self->{http}->post($url, {
        content => $payload,
        headers => { 'Content-Type' => 'application/json' },
    });

    die "HTTP error: $resp->{status} $resp->{reason}" unless $resp->{success};

    my $result = JSON::PP->new->utf8->decode($resp->{content});

    if (($result->{code} // -1) != 0) {
        die "server error code: $result->{code}";
    }

    my $changed_data = $result->{data} || {};

    if (my $ts = $changed_data->{'latest.fetch.timestamp'}) {
        $self->{last_time} = int($ts);
    }
    delete $changed_data->{'latest.fetch.timestamp'};

    # handle deleted keys
    if (my $deleted_keys_str = delete $changed_data->{deletedKeys}) {
        my @deleted = split /,/, $deleted_keys_str;
        for my $key (@deleted) {
            delete $self->{configs}{$key};
        }
        my $count = scalar @deleted;
        print "[ConfigClient] $count configs deleted: " . join(', ', @deleted) . "\n";
    }

    if (%$changed_data) {
        my %changed;
        for my $k (keys %$changed_data) {
            $self->{configs}{$k} = $changed_data->{$k};
            $changed{$k} = $changed_data->{$k};
        }
        $self->{listener}->notify(\%changed);
        my $count = scalar keys %changed;
        print "[ConfigClient] $count configs changed: " . join(', ', keys %changed) . "\n";
    }
}

sub _apply_configs {
    my ($self, $data) = @_;
    for my $k (keys %$data) {
        next if $k =~ /^kenaito\./ || $k eq 'latest.fetch.timestamp';
        next unless defined $data->{$k} && !ref $data->{$k};
        $self->{configs}{$k} = $data->{$k};
    }
}

1;
