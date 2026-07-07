use strict;
use warnings;
use ExtUtils::MakeMaker;

WriteMakefile(
    NAME         => 'Kenaito::Conf',
    VERSION_FROM => 'lib/Kenaito/Conf.pm',
    ABSTRACT     => 'Kenaito Config Center Perl Client',
    AUTHOR       => 'odboy',
    LICENSE      => 'apache_2_0',
    PREREQ_PM    => {
        'HTTP::Tiny' => '0.070',
        'JSON::PP'   => '0',
    },
    test        => { TESTS => 't/*.t' },
);
