#!/usr/bin/perl
use strict;

use WWW::Telegram::BotAPI;

my $last_update = 0;
my $start_time = time();

my $api = WWW::Telegram::BotAPI->new (
	token => '190438378:AAHVdKCSoTTDzp3_gBtYUG8r2iWxSczJJMU',
	async => 1,
);

my %COMMAND_HANDLERS = (
	ping => \&handler_ping
);

Mojo::IOLoop->recurring(0.5 => sub {
	$api->getUpdates({offset => $last_update + 1}, sub {
		my ($ua, $tx) = @_;
		die "Something bad happened!" unless $tx->success;

		my $res = $tx->res->json;
		return unless $res->{ok};

		foreach my $update (@{$res->{result}}) {
			$last_update = $update->{update_id};
			next if $update->{message}{date} < $start_time;

			my $command = $update->{message}{text};
			$command =~ s!^/!!;
			if ($COMMAND_HANDLERS{$command}) {
				$COMMAND_HANDLERS{$command}->($api, $update->{message});
			}
		}
	});
});

Mojo::IOLoop->start;

sub handler_ping {
	my ($api, $message) = @_;

	$api->sendMessage ({
		chat_id => $message->{chat}{id},
		text    => 'Pong!'
	}, sub {});
}
