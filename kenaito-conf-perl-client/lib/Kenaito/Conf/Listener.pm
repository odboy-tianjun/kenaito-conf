package Kenaito::Conf::Listener;
use strict;
use warnings;

sub new {
    my ($class) = @_;
    return bless { handlers => [] }, $class;
}

sub on_change {
    my ($self, $handler) = @_;
    push @{$self->{handlers}}, $handler;
}

sub notify {
    my ($self, $changed) = @_;
    for my $h (@{$self->{handlers}}) {
        eval { $h->($changed); 1 } or warn "[ConfigClient] change handler error: $@\n";
    }
}

1;
