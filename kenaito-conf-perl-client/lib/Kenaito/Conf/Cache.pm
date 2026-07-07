package Kenaito::Conf::Cache;
use strict;
use warnings;
use JSON::PP;
use File::Spec;

use constant CACHE_FILE => 'config-cache.json';

sub new {
    my ($class, $cache_dir) = @_;
    $cache_dir //= '.';
    return bless { dir => $cache_dir }, $class;
}

sub write {
    my ($self, $data) = @_;
    my $path = File::Spec->catfile($self->{dir}, CACHE_FILE);
    open my $fh, '>:encoding(UTF-8)', $path or return;
    print $fh JSON::PP->new->utf8->encode($data);
    close $fh;
}

sub read {
    my ($self) = @_;
    my $path = File::Spec->catfile($self->{dir}, CACHE_FILE);
    open my $fh, '<:encoding(UTF-8)', $path or return undef;
    local $/;
    my $json = <$fh>;
    close $fh;
    return eval { JSON::PP->new->utf8->decode($json) };
}

1;
