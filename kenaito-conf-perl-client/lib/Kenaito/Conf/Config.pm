package Kenaito::Conf::Config;
use strict;
use warnings;

sub new {
    my ($class, %args) = @_;
    return bless {
        base_url      => $args{base_url}      // 'http://localhost:8080',
        app_name      => $args{app_name}      // '',
        env           => $args{env}           // 'dev',
        version       => $args{version}       // 'latest',
        token         => $args{token}         // '',
        sync_interval => $args{sync_interval} // 60,
        ignore_global => $args{ignore_global} // 0,
        cache_dir     => $args{cache_dir}     // '.',
    }, $class;
}

1;
