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

use discord::{ChannelRef, Connection, Discord, State};
use discord::model::{ChannelId, LiveServer, ServerId};

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
macro_rules! unwrap_cache {
	($cache:expr) => {
		{
			if $cache.is_none() {
				fail!("Could not find in local cache.")
			}
			$cache.unwrap()
		}
	}
}
macro_rules! pretty_json {
	($($json:tt)+) => {
		{
			let json = json!($($json)+);
			let json = ::serde_json::to_string_pretty(&json);

			if json.is_err() {
				fail!("Unable to generate JSON");
			}
			json.unwrap()
		}
	}
}

// TODO!!!!
#[allow(dead_code)]
pub struct CommandContext {
	pub session: Discord,
	pub websocket: Connection,
	pub state: State,

	pub guild: Option<ServerId>,
	pub channel: Option<ChannelId>
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

// Shut clippy up about my macros... for now at least
#[cfg_attr(feature = "cargo-clippy", allow(needless_return))]
#[cfg_attr(feature = "cargo-clippy", allow(deref_addrof))]
pub fn execute(context: &mut CommandContext, tokens: &[String]) -> CommandResult {
	if tokens.len() < 1 {
		return CommandResult {
		           empty: true,
		           ..Default::default()
		       };
	}
	let command = &tokens[0];
	let tokens = &tokens[1..];
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
