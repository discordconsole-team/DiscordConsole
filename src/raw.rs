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
extern crate rustyline;

use self::rustyline::Editor;
use self::rustyline::error::ReadlineError;
use color::*;
use command::CommandContext;
use discord::ChannelRef;
use std::sync::{Arc, Mutex};

pub fn raw(context: Arc<Mutex<CommandContext>>) {
	let mut rl = Editor::<()>::new();

	loop {
		// println!("Pointer: Locking {:?}", context);
		let prefix = {
			pointer(&context.lock().unwrap(), true)
		};
		// println!("Pointer: Unlocked {:?}", context);
		let prefix = &prefix;

		let mut first = true;
		let mut command = String::new();

		let tokens = ::tokenizer::tokens(|| {
			let wasfirst = first;
			first = false;

			let result = rl.readline(if wasfirst { prefix } else { "" });

			match result {
				Ok(res) => {
					if !wasfirst {
						command.push(' ');
					}
					command.push_str(&res);

					Ok(res)
				},
				Err(err) => Err(err),
			}
		});
		rl.add_history_entry(&command);
		let tokens = match tokens {
			Ok(tokens) => tokens,
			Err(ReadlineError::Eof) |
			Err(ReadlineError::Interrupted) => {
				break;
			},
			Err(err) => {
				eprintln!("Error reading line: {}", err);
				break;
			},
		};

		// println!("Command: Locking");
		let result = ::command::execute(&mut context.lock().unwrap(), true, tokens);
		// println!("Command: Executed");
		if result.success {
			if let Some(text) = result.text {
				println!("{}", &text);
			}
		} else if let Some(text) = result.text {
			eprintln!("{}{}{}", *COLOR_RED, &text, *COLOR_RESET);
		}

		// println!("Command: Unlocked");
		if result.exit {
			break;
		}
	}
}

pub fn pointer(context: &CommandContext, terminal: bool) -> String {
	let mut capacity = 2; // Minimum capacity
	if terminal {
		capacity += COLOR_YELLOW.len();
		capacity += COLOR_RESET.len();
	}

	let mut prefix = String::with_capacity(capacity);
	if terminal {
		prefix.push_str(if context.using.is_some() {
			*COLOR_CYAN
		} else {
			*COLOR_YELLOW
		});
	}
	if let Some(guild) = context.guild {
		prefix.push_str(match context.state.find_server(guild) {
			Some(guild) => &guild.name,
			None => "Unknown",
		});
	}
	if let Some(channel) = context.channel {
		prefix.push_str(" (");
		prefix.push_str(&match context.state.find_channel(channel) {
			Some(channel) => {
				match channel {
					ChannelRef::Public(_, channel) => {
						let mut name = channel.name.clone();
						name.insert(0, '#');
						name
					},
					ChannelRef::Group(channel) => channel.name.clone().unwrap_or_default(),
					ChannelRef::Private(channel) => channel.recipient.name.clone(),
				}
			},
			None => "unknown".to_string(),
		});
		prefix.push_str(")");
	}
	prefix.push_str("> ");
	if terminal {
		prefix.push_str(*COLOR_RESET);
	}
	prefix
}
