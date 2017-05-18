extern crate clap;
use self::clap::{App, Arg};

use std::io::Write;

#[derive(Debug)]
pub struct Options {
	pub tokens: Vec<String>,
	pub token: usize,

	pub notui: bool
}

pub fn get_options() -> Options {
	let args = App::new("discord_console")
		.version(super::VERSION)
		.about("Use discord in a new way")
		.author("LEGOlord208")
		.arg(
			Arg::with_name("token")
				.long("token")
				.short("t")
				.help("Specify Discord token")
				.multiple(true)
				.takes_value(true)
		)
		.arg(
			Arg::with_name("notui")
				.long("notui")
				.help("No Text UI will be used. Pure command mode.")
		)
		.get_matches();

	let tokens = args.values_of("token");
	let tokens = if tokens.is_some() {
		tokens.unwrap().map(|s| s.to_string()).collect()
	} else {
		print!("Token: ");
		flush!();

		let mut token = String::new();
		super::std::io::stdin().read_line(&mut token).unwrap();

		vec![token]
	};

	Options {
		tokens: tokens,
		token: 0,

		notui: args.is_present("notui")
	}
}
