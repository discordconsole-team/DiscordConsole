/* DiscordConsole is a software aiming to give you full control over
 * accounts, bots and webhooks!
 * Copyright (C) 2017  LEGOlord208
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 * */
extern crate rustyline;

use self::rustyline::Editor;
use self::rustyline::error::ReadlineError;
use color::*;

use discord::ChannelRef;
use std::io::Write;

pub fn raw(mut context: ::command::CommandContext) {
	let mut rl = Editor::<()>::new();

	loop {
		let mut prefix = String::with_capacity(COLOR_YELLOW.len() + 2 + COLOR_RESET.len()); // Minimum capacity
		prefix.push_str(*COLOR_YELLOW);
		if let Some(guild) = context.guild {
			prefix.push_str(
				match ::command::find_guild(&context.state, guild) {
					Some(guild) => guild.name.as_str(),
					None => "Unknown",
				}
			);
		}
		if let Some(channel) = context.channel {
			prefix.push_str(" (");
			prefix.push_str(
				match context.state.find_channel(channel) {
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
					}
					.as_str()
			);
			prefix.push_str(")");
		}
		prefix.push_str("> ");
		prefix.push_str(*COLOR_RESET);
		let prefix = prefix.as_str();

		let mut first = true;
		let mut command = String::new();

		let tokens = ::tokenizer::tokens(
			|| {
				let result = rl.readline(
					if first {
						first = false;
						prefix
					} else {
						""
					}
				);

				match result {
					Ok(res) => {
						command.push_str(res.as_str());
						Ok(res)
					},
					Err(err) => Err(err),
				}
			}
		);
		rl.add_history_entry(command.as_str());
		let tokens = match tokens {
			Ok(tokens) => tokens,
			Err(ReadlineError::Eof) |
			Err(ReadlineError::Interrupted) => {
				break;
			},
			Err(err) => {
				stderr!("Error reading line: {}", err);
				break;
			},
		};

		let result = ::command::execute(&mut context, &tokens);
		if result.success {
			if let Some(text) = result.text {
				println!("{}", text.as_str());
			}
		} else if let Some(text) = result.text {
			stderr!("{}", text.as_str());
		}

		if result.exit {
			break;
		}
	}
}
