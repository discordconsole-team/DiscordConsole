// DiscordConsole is a software aiming to give you full control over
// accounts, bots and webhooks!
// Copyright (C) 2017  LEGOlord208
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
extern crate clap;
use self::clap::{App, Arg};
use color::*;

use std::io::Write;

#[derive(Debug)]
pub struct Options {
	pub tokens: Vec<String>,
	pub token: usize,

	pub notui: bool
}

pub fn get_options() -> Option<Options> {
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
		.arg(Arg::with_name("notui").long("notui").help(
			"No Text UI will be used. Pure command mode."
		))
		.get_matches();

	let tokens = args.values_of("token");
	let tokens = if tokens.is_some() {
		tokens.unwrap().map(|s| s.to_string()).collect()
	} else {
		println!(
			"To use your account, we need your {}token{}.",
			*COLOR_YELLOW,
			*COLOR_RESET
		);
		println!("'tokens' in Discord and a few other services");
		println!("represent your session.");
		println!("They're what keeps you logged in.");
		println!(
			"{}Your token is re-generated when you change password{}",
			*COLOR_RED,
			*COLOR_RESET
		);
		println!();
		println!("Care to let us highjack your session?");
		print!("Token: ");
		flush!();

		let mut token = String::new();
		match super::std::io::stdin().read_line(&mut token) {
			Ok(n) => {
				if n <= 0 {
					return None;
				}
			},
			Err(err) => {
				stderr!("Couldn't read line: {}", err);
				return None;
			},
		}

		vec![token]
	};

	Some(Options {
		tokens: tokens,
		token: 0,

		notui: args.is_present("notui")
	})
}
