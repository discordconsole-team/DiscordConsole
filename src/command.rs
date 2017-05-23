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


use color::*;
use discord::{ChannelRef, Connection, Discord, State};
use discord::model::{ChannelId, ChannelType, LiveServer, ServerId};

use std::collections::HashMap;
use std::fs::File;
use std::io::{BufRead, BufReader};
use std::process::{Command, Stdio};

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
macro_rules! to_id {
	($type:expr, $context:expr, $funcid:ident, $funcname:ident, $ref:expr, $nameorid:expr) => {
		{
			let i = $nameorid.parse();
			let mut val;

			if i.is_err() {
				val = $context.state.$funcname($context.guild, $nameorid.as_str())
			} else {
				val = $context.state.$funcid($type(i.unwrap()));
				if val.is_none() {
					val = $context.state.$funcname($context.guild, $nameorid.as_str())
				}
			}

			val
		}
	}
}
macro_rules! attempt {
	($result:expr, $message:expr) => {
		{
			if $result.is_err() {
				fail!($message);
			}

			$result.unwrap()
		}
	}
}
macro_rules! require {
	($option:expr, $message:expr) => {
		{
			if $option.is_none() {
				fail!($message);
			}

			$option.unwrap()
		}
	}
}
macro_rules! unwrap_cache {
	($cache:expr) => {
		require!($cache, "Could not find in local cache.")
	}
}
macro_rules! pretty_json {
	($($json:tt)+) => {
		{
			let json = json!($($json)+);
			let json = ::serde_json::to_string_pretty(&json);

			attempt!(json, "Unable to generate JSON.")
		}
	}
}
macro_rules! require_guild {
	($context:expr) => {
		require!($context.guild, "This command requires a guild to be selected.")
	}
}
macro_rules! require_channel {
	($context:expr) => {
		require!($context.channel, "This command requires a channel to be selected.")
	}
}

pub struct CommandContext {
	pub session: Discord,
	pub websocket: Connection,
	pub state: State,

	pub guild: Option<ServerId>,
	pub channel: Option<ChannelId>,

	pub alias: HashMap<String, Vec<String>>,
	pub terminal: bool
}
impl CommandContext {
	pub fn new(session: Discord, websocket: Connection, state: State) -> CommandContext {
		CommandContext {
			session: session,
			websocket: websocket,
			state: state,

			guild: None,
			channel: None,

			alias: HashMap::new(),
			terminal: false
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

// Shut clippy up about my macros... for now at least
#[cfg_attr(feature = "cargo-clippy", allow(needless_return))]
pub fn execute(context: &mut CommandContext, mut tokens: Vec<String>) -> CommandResult {
	if tokens.len() < 1 {
		return CommandResult {
		           empty: true,
		           ..Default::default()
		       };
	}
	let mut command = tokens[0].clone();
	tokens.remove(0);

	// Unsure about the best approach here.
	// Used to take a slice to this whole function, but it'd cause issues
	// when this came along...
	if let Some(atokens) = context.alias.get(&command) {
		let mut atokens = atokens.clone();

		command = atokens[0].clone();
		atokens.remove(0);
		atokens.append(&mut tokens);
		tokens = atokens;
	}

	let command = command.as_str();

	match command {
		"echo" => {
			usage_one!(tokens, "echo <text>");
			success!(Some(tokens[0].clone()));
		},
		"alias" => {
			match tokens.len() {
				0 => {
					let mut output = String::new();
					let mut first = true;

					for (key, val) in &context.alias {
						if first {
							first = false;
						} else {
							output.push('\n');
						}
						output.push_str("alias ");
						output.push_str(key.as_str());
						output.push(' ');
						output.push_str(::escape::escape(&val).as_str());
					}

					success!(
						if output.is_empty() {
							None
						} else {
							Some(output)
						}
					);
				},
				1 => {
					context.alias.remove(tokens[0].as_str());
					success!(None);
				},
				_ => {
					let name = tokens[0].clone();
					if name == "alias" {
						fail!("lol nope");
					}
					context.alias.insert(name, tokens[1..].to_vec());

					success!(None);
				},
			}
		},
		"exec" => {
			usage_min!(tokens, 2, "exec <type> <command>");

			match tokens[0].as_str() {
				"shell" => {
					usage_max!(tokens, 2, "exec shell <command>");

					let cmd = if cfg!(target_os = "windows") {
						Command::new("cmd")
							.arg("/c")
							.arg(tokens[1].clone())
							.stdin(Stdio::inherit())
							.stdout(Stdio::inherit())
							.stderr(Stdio::inherit())
							.status()
					} else {
						Command::new("sh")
							.arg("-c")
							.arg(tokens[1].clone())
							.stdin(Stdio::inherit())
							.stdout(Stdio::inherit())
							.stderr(Stdio::inherit())
							.status()

					};
					if cmd.is_err() {
						fail!("Could not execute command");
					}
					success!(
						Some(
							format!(
								"{}Process exited with status {}{}",
								if context.terminal { *COLOR_BLACK } else { "" },
								cmd.unwrap().code().unwrap_or(1),
								if context.terminal { *COLOR_RESET } else { "" },
							)
						)
					);
				},
				"file" => {
					usage_max!(tokens, 2, "exec file <file>");

					let file = attempt!(File::open(tokens[1].clone()), "Could not open file");
					let bufreader = BufReader::new(file);

					let mut results = String::new();
					let mut first = true;

					for line in bufreader.lines() {
						let line = attempt!(line, "Could not read line from file");

						if first {
							first = false;
						} else {
							results.push('\n');
						}

						let mut first = true;
						let tokens = ::tokenizer::tokens(
							|| if first {
								first = false;
								Ok(line.clone())
							} else {
								Err(())
							}
						);
						let tokens = attempt!(tokens, "Unclosed quote or trailing \\");
						let result = execute(context, tokens);

						if result.empty {
							continue;
						}
						if result.exit {
							return CommandResult {
							           exit: true,
							           ..Default::default()
							       };
						}

						let mut prefix = ::raw::pointer(context);
						if context.terminal {
							prefix.push_str(*COLOR_ITALIC);
						}
						prefix.push_str(line.as_str());
						if context.terminal {
							prefix.push_str(*COLOR_RESET);
						}
						prefix.push('\n');
						results.push_str(prefix.as_str());
						results.push_str(result.text.unwrap_or_default().as_str())
					}

					success!(Some(results));
				},
				_ => fail!("Not a valid type."),
			}
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

			if tokens.is_empty() {
				context.guild = None;
				context.channel = None;
				success!(None);
			}
			let guild = to_id!(
				ServerId,
				context,
				find_guild,
				find_guild_by_name,
				&mut guild,
				tokens[0]
			);

			let guild = unwrap_cache!(guild);
			context.guild = Some(guild.id);
			context.channel = Some(guild.id.main());

			success!(
				Some(
					pretty_json!({
						"id":       guild.id.to_string().as_str(),
						"name":     guild.name.as_str(),
						"owner_id": guild.owner_id.to_string().as_str(),
					})
				)
			);
		},
		"channel" => {
			usage_max!(tokens, 1, "channel <id/name>");

			if tokens.is_empty() {
				if let Some(guild) = context.guild {
					context.channel = Some(guild.main());
				} else {
					context.channel = None;
				}
				success!(None);
			}
			let channel = to_id!(
				ChannelId,
				context,
				find_channel,
				find_channel_by_name,
				&mut channel,
				tokens[0]
			);
			let channel = unwrap_cache!(channel);

			match channel {
				ChannelRef::Private(channel) => {
					context.guild = None;
					context.channel = Some(channel.id);

					success!(Some(pretty_json!({
						"id":        channel.id.to_string().as_str(),
						"recipient": {
							"id":   channel.recipient.id.to_string().as_str(),
							"name": channel.recipient.name.as_str()
						}
					})));
				},
				ChannelRef::Group(channel) => {
					context.guild = None;
					context.channel = Some(channel.channel_id);

					success!(Some(pretty_json!({
						"id":       channel.channel_id.to_string().as_str(),
						"name":     channel.name.clone().unwrap_or_default().as_str()
					})));
				},
				ChannelRef::Public(guild, channel) => {
					context.guild = Some(guild.id);
					context.channel = Some(channel.id);

					success!(Some(pretty_json!({
						"id":       channel.id.to_string().as_str(),
						"name":     channel.name.as_str(),
						"guild": {
							"id":   guild.id.to_string().as_str(),
							"name": guild.name.as_str()
						}
					})));
				},
			}
		},
		"guilds" => {
			usage_max!(tokens, 0, "guilds");

			let mut guilds = context.state.servers().to_vec();
			if let Some(settings) = context.state.settings() {
				::sort::sort_guilds(settings, &mut guilds);
			}

			let mut value = String::new();
			let mut first = true;
			for guild in guilds {
				if first {
					first = false;
				} else {
					value.push('\n');
				}
				value.push_str(guild.id.to_string().as_str());
				value.push(' ');
				value.push_str(guild.name.as_str());
			}

			success!(Some(value));
		},
		"channels" => {
			usage_max!(tokens, 0, "channels");
			let guild = require_guild!(context);
			let guild = unwrap_cache!(context.state.find_guild(guild));

			let mut value = String::new();
			let mut first = true;

			for kind in [ChannelType::Text, ChannelType::Voice].iter() {
				let mut channels = guild.channels.iter().filter(|x| x.kind == *kind).collect();
				::sort::sort_channels(&mut channels);

				for channel in channels {
					if first {
						first = false;
					} else {
						value.push('\n');
					}
					value.push_str(channel.id.to_string().as_str());
					value.push(' ');
					value.push_str(channel.kind.name());
					value.push(' ');
					value.push_str(channel.name.as_str());
				}
			}

			success!(Some(value));
		},
		"say" => {
			usage_max!(tokens, 1, "say [text]");
			// TODO :^)
			success!(None);
		},
		_ => {
			fail!("Unknown command!");
		},
	}
}

pub trait MoreStateFunctionsSuperOriginalTraitNameExclusiveTM {
	fn find_guild(&self, id: ServerId) -> Option<&LiveServer>;
	fn find_guild_by_name<'a>(&'a self, guild: Option<ServerId>, name: &str) -> Option<&'a LiveServer>;
	fn find_channel_by_name<'a>(&'a self, guild: Option<ServerId>, name: &'a str) -> Option<ChannelRef<'a>>;
}
impl MoreStateFunctionsSuperOriginalTraitNameExclusiveTM for State {
	fn find_guild(&self, id: ServerId) -> Option<&LiveServer> {
		for guild in self.servers() {
			if guild.id == id {
				return Some(guild);
			}
		}
		None
	}

	// Unsure what the best way to deal with this is.
	// The function is called from a macro.
	#[allow(unused_variables)]
	fn find_guild_by_name<'a>(&'a self, guild: Option<ServerId>, name: &str) -> Option<&'a LiveServer> {
		for guild in self.servers() {
			if guild.name == name {
				return Some(guild);
			}
		}
		None
	}

	fn find_channel_by_name<'a>(&'a self, guild: Option<ServerId>, name: &str) -> Option<ChannelRef<'a>> {
		for guild2 in self.servers() {
			if guild.is_some() && guild2.id != guild.unwrap() {
				continue;
			}
			for channel in &guild2.channels {
				if channel.name == name {
					return Some(ChannelRef::Public(guild2, channel));
				}
			}
		}
		let some_name = Some(name.to_string());
		for group in self.groups().values() {
			if group.name == some_name {
				return Some(ChannelRef::Group(group));
			}
		}
		for private in self.private_channels() {
			if private.recipient.name == name {
				return Some(ChannelRef::Private(private));
			}
		}
		None
	}
}
