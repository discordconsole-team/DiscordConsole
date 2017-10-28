// DiscordConsole is a software aiming to give you full control over
// accounts, bots and webhooks!
// Copyright (C) 2017  jD91mZM2
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
		let prompt = {
			prompt(&context.lock().unwrap(), true)
		};
		let prompt = &prompt;

		let mut first = true;
		let mut command = String::new();

		let tokens = ::parser::parse(|| {
			let wasfirst = first;
			first = false;

			let result = rl.readline(if wasfirst { prompt } else { "" });

			match result {
				Ok(result) => {
					if !wasfirst {
						command.push(' ');
					}
					command.push_str(&result);

					Ok(result)
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

pub fn prompt(context: &CommandContext, terminal: bool) -> String {
	let mut capacity = 0; // Color capacity
	if terminal {
		capacity += COLOR_YELLOW.len();
		capacity += COLOR_RESET.len();
	}

	let mut result = String::with_capacity(capacity);
	if terminal {
		result.push_str(if context.using.is_some() {
			*COLOR_CYAN
		} else {
			*COLOR_YELLOW
		});
	}
	let guild = match context.guild {
		Some(guild) => {
			match context.state.find_server(guild) {
				Some(guild) => guild.name.clone(),
				None => "Unknown".to_string(),
			}
		},
		None => "".to_string(),
	};
	let channel = match context.channel {
		Some(channel) => {
			match context.state.find_channel(channel) {
				Some(channel) => {
					match channel {
						ChannelRef::Public(_, channel) => {
							let mut name = String::with_capacity(channel.name.len() + 1);
							name.push('#');
							name.push_str(&channel.name);
							name
						},
						ChannelRef::Group(channel) => channel.name.clone().unwrap_or_default(),
						ChannelRef::Private(channel) => channel.recipient.name.clone(),
					}
				},
				None => "unknown".to_string(),
			}
		},
		None => "<no channel>".to_string(),
	};

	let format_str = match (context.guild, context.channel) {
		(None, None) => &context.ptr0,
		(None, Some(_)) => &context.ptr1,
		(Some(_), None) |
		(Some(_), Some(_)) => &context.ptr2,
	};

	result.reserve(format_str.len());

	let mut escape = false;
	for c in format_str.chars() {
		if escape {
			escape = false;

			let expanded = match c {
				'%' => "%",
				'g' => &guild,
				'c' => &channel,
				'e' => "\x1b",
				_ => {
					result.push('%');
					result.push(c);
					continue;
				},
			};
			result.push_str(expanded);
		} else if c == '%' {
			escape = true;
		} else {
			result.push(c);
		}
	}

	if terminal {
		result.push_str(*COLOR_RESET);
	}
	result
}
