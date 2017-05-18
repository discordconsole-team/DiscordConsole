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

use discord::{Connection, Discord, State};

macro_rules! success {
	($val:expr) => {
		return CommandResult{
			text:    $val,
			..Default::default()
		}
	}
}
macro_rules! fail {
	($val:expr) => {
		return CommandResult{
			text:    Some($val.to_string()),
			success: false,
			..Default::default()
		}
	}
}
macro_rules! usage_min {
	($tokens:expr, $min:expr, $usage:expr) => {
		if $tokens.len() < $min {
			fail!(concat!($usage, "\nYou supplied too few arguments."));
		}
	}
}
macro_rules! usage_max {
	($tokens:expr, $max:expr, $usage:expr) => {
		if $tokens.len() > $max {
			fail!(concat!($usage, "\nYou supplied too many arguments."));
		}
	}
}
macro_rules! usage {
	($tokens:expr, $exact:expr, $usage:expr) => {
		usage_min!($tokens, $exact, $usage);
		usage_max!($tokens, $exact, $usage);
	}
}
macro_rules! usage_one {
	($tokens:expr, $usage:expr) => {
		if $tokens.len() != 1 {
			fail!(concat!($usage, "\nYou did not supply 1 argument.\n\
							Did you mean to put quotes around the argument?"));
		}
	}
}

// TODO!!!!
#[allow(dead_code)]
pub struct CommandContext {
	pub session: Discord,
	pub websocket: Connection,
	pub state: State,

	guild: Option<String>,
	channel: Option<String>
}
impl CommandContext {
	pub fn new(session: Discord, conn: Connection, state: State) -> CommandContext {
		CommandContext {
			session: session,
			websocket: conn,
			state: state,

			guild: None,
			channel: None
		}
	}
}
pub struct CommandResult {
	pub text: Option<String>,
	pub success: bool,
	pub exit: bool,
	pub empty: bool
}
impl Default for CommandResult {
	fn default() -> CommandResult {
		CommandResult {
			text: None,
			success: true,
			exit: false,
			empty: false
		}
	}
}

pub fn execute(context: &mut CommandContext, mut tokens: Vec<String>) -> CommandResult {
	if tokens.len() < 1 {
		return CommandResult {
		           empty: true,
		           ..Default::default()
		       };
	}
	let command = tokens.remove(0);
	let command = command.as_str();

	match command {
		"echo" => {
			usage_one!(tokens, "echo <text>");
			success!(Some(tokens[0].clone()));
		},
		"exit" => {
			usage_max!(tokens, 0, "exit");
			CommandResult {
				exit: true,
				..Default::default()
			}
		},
		"guild" => {
			usage_max!(tokens, 1, "guild <id/name>");
			context.guild = if tokens.len() < 1 {
				None
			} else {
				Some(tokens[0].clone())
			};
			success!(None);
		},
		_ => {
			fail!("Unknown command!");
		},
	}
}
