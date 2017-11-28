#!/usr/bin/env perl
use strict;
use warnings;
use 5.018;

use File::Basename;

my $dir = "$ENV{GOPATH}/src/github.com/linuxdeepin/go-gir";
say $dir;

my @dirs = glob "$dir/*";

for my $d (@dirs) {
	if (! -d $d) {
		next
	}

	my $bname = basename($d);
	if ($bname =~ /(\w+)-([\d.]+)/) {
		say $d;
		system './try-error', '-list-only', $d;
	}
}
