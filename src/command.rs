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

extern crate hlua;

use self::hlua::{AnyLuaValue, Lua};
use color::*;
use discord::{ChannelRef, Connection, Discord, State};
use discord::model::{ChannelId, ChannelType, LiveServer, MessageId, ServerId};
use escape::escape;
use std::cmp;
use std::collections::HashMap;
use std::error::Error;
use std::fmt;
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
			fail!(concat!($usage, "\nYou supplied too few arguments"));
		}
	}
}
macro_rules! usage_max {
	($tokens:expr, $max:expr, $usage:expr) => {
		if $tokens.len() > $max {
			fail!(concat!($usage, "\nYou supplied too many arguments"));
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
macro_rules! from_id {
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
			if let Err(err) = $result {
				fail!(format!("{} (Details: {})", $message, err));
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
		require!($cache, couldnt!("find in local cache"))
	}
}
macro_rules! pretty_json {
	($($json:tt)+) => {
		{
			let json = json!($($json)+);
			let json = ::serde_json::to_string_pretty(&json);

			attempt!(json, "Unable to generate JSON")
		}
	}
}
macro_rules! require_guild {
	($context:expr) => {
		require!($context.guild, "This command requires a guild to be selected")
	}
}
macro_rules! require_channel {
	($context:expr) => {
		require!($context.channel, "This command requires a channel to be selected")
	}
}
macro_rules! unknown {
	($what:expr) => {
		{ concat!("Unknown ", $what) }
	}
}
macro_rules! couldnt {
	($what:expr) => {
		{ concat!("Could not ", $what) }
	}
}
macro_rules! parse {
	($str:expr) => {
		{
			let num = $str.parse();
			attempt!(num, "Not a number")
		}
	}
}

pub struct CommandContext {
	pub session: Discord,
	pub websocket: Connection,
	pub state: State,

	pub guild: Option<ServerId>,
	pub channel: Option<ChannelId>,

	pub terminal: bool,

	pub alias: HashMap<String, Vec<String>>,
	pub using: Option<Vec<String>>
}
impl CommandContext {
	pub fn new(session: Discord, websocket: Connection, state: State) -> CommandContext {
		CommandContext {
			session: session,
			websocket: websocket,
			state: state,

			guild: None,
			channel: None,

			terminal: false,

			alias: {
				let mut map = HashMap::new();
				map.insert(
					"say".to_string(),
					vec!["msg".to_string(), "normal".to_string(), "send".to_string()]
				);
				map.insert(
					"tts".to_string(),
					vec!["msg".to_string(), "tts".to_string(), "send".to_string()]
				);
				map.insert(
					"embed".to_string(),
					vec!["msg".to_string(), "embed".to_string(), "send".to_string()]
				);
				map.insert(
					"edit".to_string(),
					vec!["msg".to_string(), "normal".to_string()]
				);

				map
			},
			using: None
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

#[cfg_attr(feature = "cargo-clippy", allow(needless_return))]
// Shut clippy up about my macros... for now at least
#[cfg_attr(feature = "cargo-clippy", allow(cyclomatic_complexity))]
// Unsure if I really should split it up. It shall be thought about.
pub fn execute(context: &mut CommandContext, mut tokens: Vec<String>) -> CommandResult {
	if tokens.len() < 1 {
		if context.using.is_some() {
			context.using = None;
		}
		return CommandResult {
		           empty: true,
		           ..Default::default()
		       };
	}

	// Unsure about the best approach here.
	// Used to take a slice to this whole function, but it'd cause issues
	// when these came along...
	if let Some(ref using) = context.using {
		let mut using = using.clone();

		using.append(&mut tokens);
		tokens = using;
	}
	if let Some(atokens) = context.alias.get(&tokens[0]) {
		let mut atokens = atokens.clone();

		tokens.remove(0);
		atokens.append(&mut tokens);
		tokens = atokens;
	}
	let command = tokens[0].clone();
	tokens.remove(0);
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
						output.push_str(escape(key).as_str());
						output.push_str(" = ");
						output.push_str(
							val.iter()
								.map(|item| escape(item))
								.collect::<Vec<String>>()
								.join(" ")
								.as_str()
						);
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

					let start = if tokens[1] == "=" { 2 } else { 1 };
					context.alias.insert(name, tokens[start..].to_vec());

					success!(None);
				},
			}
		},
		"exec" => {
			usage_min!(tokens, 2, "exec <type> <value>");

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
						fail!(couldnt!("execute command"));
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
					let result = execute_file(context, tokens[1].clone());
					let result = attempt!(result, couldnt!("run commands file"));

					success!(Some(result))
				},
				"lua" => {
					let mut lua = new_lua(context);

					let file = attempt!(File::open(tokens[1].clone()), couldnt!("open file"));
					if let Err(err) = lua.execute_from_reader::<(), _>(file) {
						fail!(format!("Error trying to execute: {:?}", err));
					}
					success!(None);
				},
				"lua-inline" => {
					usage_max!(tokens, 2, "exec lua-inline <text>");
					let mut lua = new_lua(context);

					if let Err(err) = lua.execute::<()>(tokens[1].clone().as_str()) {
						fail!(format!("Error trying to execute: {:?}", err));
					}
					success!(None);
				},
				_ => fail!(unknown!("type (shell/file/lua/lua-inline available)")),
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
			let guild = from_id!(
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
			let channel = from_id!(
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

			for kind in &[ChannelType::Text, ChannelType::Voice] {
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
		"msg" => {
			usage!(tokens, 3, "msg <type> <\"send\"/existing id> <text>");
			let channel = require_channel!(context);

			let kind = match tokens[0].clone().as_str() {
				"normal" => 0,
				"tts" => 1,
				"embed" => 2,
				_ => fail!(unknown!("type (normal/tts/embed available)")),
			};
			let edit = match tokens[1].clone().as_str() {
				"send" => None,
				id => Some(parse!(id)),
			};

			let text = tokens[2].clone();
			let mut text = text.as_str();

			let mut output = String::new();
			let mut first = true;

			while !text.is_empty() {
				if first {
					first = false;
				} else {
					output.push('\n');
				}
				let amount = cmp::min(text.len(), ::LIMIT_MSG);
				let value = &text[..amount];
				text = &text[amount..];

				let msg = match kind {
					0 => {
						if let Some(edit) = edit {
							context
								.session
								.edit_message(channel, MessageId(edit), value)
						} else {
							context.session.send_message(channel, value, "", false)
						}
					},
					1 => {
						if edit.is_some() {
							fail!("Can't edit TTS");
						}
						context.session.send_message(channel, value, "", true)
					},
					2 => {
						fail!("Not implemented. Waiting for discord-rs. See https://github.com/SpaceManiac/discord-rs/issues/112");
						/*
						if context
						       .session
						       .send_embed(channel, value, |builder| builder.description("Hi"))
						       .is_err() {
							fail!(couldnt!("send embed"));
						}
						*/
					},
					_ => unreachable!(),
				};
				let msg = attempt!(msg, couldnt!("send message"));
				output.push_str(format!("Sent message with ID {}", msg.id).as_str());
			}

			success!(Some(output));
		},
		"use" => {
			usage_min!(tokens, 1, "use <command...>");

			context.using = Some(tokens);
			success!(Some("Use mode enabled.\nSend an empty command to disable.".to_string()));
		},
		_ => fail!(unknown!("command")),
	}
}

#[derive(Debug)]
struct ErrUnclosed;

impl Error for ErrUnclosed {
	fn description(&self) -> &str { "Command not closed; Quote unclosed or trailing \\" }
}
impl fmt::Display for ErrUnclosed {
	fn fmt(&self, fmt: &mut fmt::Formatter) -> fmt::Result { write!(fmt, "{}", self.description()) }
}

pub fn execute_file(context: &mut CommandContext, file: String) -> Result<String, Box<Error>> {
	let file = File::open(file)?;
	let bufreader = BufReader::new(file);

	let pointer = ::raw::pointer(context);

	let mut results = String::new();
	let mut first = true;

	for line in bufreader.lines() {
		let line = line?;

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
				Err(ErrUnclosed)
			}
		);
		let tokens = tokens?;
		let result = execute(context, tokens);

		if result.empty {
			continue;
		}
		if result.exit {
			results.push_str("Can't exit from a commands file");
			continue;
		}

		results.push_str(pointer.clone().as_str());
		if context.terminal {
			results.push_str(*COLOR_ITALIC);
		}
		results.push_str(line.as_str());
		if context.terminal {
			results.push_str(*COLOR_RESET);
		}
		results.push('\n');
		results.push_str(result.text.unwrap_or_default().as_str())
	}

	Ok(results)
}

fn lua_to_string(value: AnyLuaValue) -> String {
	match value {
		AnyLuaValue::LuaString(value) => value,
		AnyLuaValue::LuaNumber(value) => (value.round() as u64).to_string(),
		AnyLuaValue::LuaBoolean(value) => value.to_string(),
		AnyLuaValue::LuaArray(value) => {
			value
				.iter()
				.map(
					|value| {
						let value0 = lua_to_string(value.0.clone());
						let value1 = lua_to_string(value.1.clone());
						let mut string = String::with_capacity(value0.len() + 2 + value1.len());
						string.push_str(value0.as_str());
						string.push_str(": ");
						string.push_str(value1.as_str());

						string
					}
				)
				.collect::<Vec<_>>()
				.join(", ")
		},
		_ => String::new(),
	}

}
pub fn new_lua(context: &mut CommandContext) -> Lua {
	let mut lua = Lua::new();
	lua.openlibs();

	// Example: `cmd({"echo", "Hello World"})`
	// crashes on incorrect type; see https://github.com/tomaka/hlua/issues/149
	lua.set(
		"cmd",
		hlua::function1::<_, String, Vec<AnyLuaValue>>(
			move |args| {
				let args = args.iter()
					.map(|value| lua_to_string(value.clone()))
					.collect();
				execute(context, args).text.unwrap_or_default()
			}
		)
	);

	lua
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
