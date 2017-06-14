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
#[macro_use]
extern crate clap;
extern crate discord;
#[macro_use]
extern crate lazy_static;
#[macro_use]
extern crate serde_json;
extern crate clipboard;

#[macro_export]
macro_rules! stderr {
	($fmt:expr)              => { writeln!(::std::io::stderr(), concat!("{}", $fmt, "{}"), *COLOR_RED, *COLOR_RESET).unwrap(); };
	($fmt:expr, $($arg:tt)*) => { writeln!(::std::io::stderr(), concat!("{}", $fmt, "{}"), *COLOR_RED, $($arg)*, *COLOR_RESET).unwrap(); };
}
macro_rules! flush {
	() => { ::std::io::stdout().flush().unwrap(); }
}

mod color;
mod command;
mod escape;
mod help;
mod options;
mod raw;
mod sort;
mod tokenizer;
mod tui;

use color::*;
use command::CommandContext;
use discord::{Connection, Discord, State};
use std::io::Write;
use std::sync::{Arc, Mutex};
// use std::thread;

const LIMIT: u16 = 100;
const LIMIT_MSG: u16 = 2000;

fn main() {
	let options = options::get_options();

	if options.is_none() {
		return;
	}
	let mut options = options.unwrap();

	for token in &mut options.tokens {
		*token = token.trim().to_string();

		if token.is_empty() {
			println!("Token cannot be empty");
			return;
		}

		let lower = token.to_lowercase();

		if lower.starts_with("bot ") {
			*token = "Bot ".to_string() + &token[4..];
		} else if lower.starts_with("user ") {
			*token = token[5..].to_string();
		}
	}

	let context = CommandContext::new(options.tokens, 0);
	if let Err(err) = context {
		stderr!("Could not connect to gateway: {}", err);
		return;
	}
	let context = Arc::new(Mutex::new(context.unwrap()));

	// TODO.
	// See https://krake.one/l/20kh
	//
	// let clone = context.clone();
	// thread::spawn(
	// 	move || loop {
	// 		println!("Event: Locking {:?}", clone);
	// 		let mut gateway = {
	// 			&mut clone.lock().unwrap().gateway
	// 		};
	// 		println!("Event: Unlocked {:?}", clone);
	// 		match gateway.recv_event() {
	// 			Ok(event) => {
	// 				// println!("Updating state: {:?}", event);
	// 				clone.lock().unwrap().state.update(&event)
	// 			},
	// 			Err(err) => {
	// 				stderr!("Error receiving: {}", err);
	// 			},
	// 		}
	// 	}
	// );

	if options.notui {
		raw::raw(context);
	} else {
		tui::tui(context);
	}

	print!("{}", *COLOR_RESET);
}

pub fn connect(token: &str) -> Result<(Discord, Connection, State), discord::Error> {
	let session = Discord::from_user_token(token).unwrap();
	let (gateway, ready) = match session.connect() {
		Ok((gateway, ready)) => (gateway, ready),
		Err(err) => {
			return Err(err);
		},
	};
	let state = State::new(ready);

	Ok((session, gateway, state))
}
