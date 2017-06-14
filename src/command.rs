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
extern crate hlua;

use self::hlua::{AnyLuaValue, Lua};
use {LIMIT, LIMIT_MSG};
use color::*;
use discord::{ChannelRef, Connection, Discord, GetMessages, State};
use discord::model::{ChannelId, ChannelType, LiveServer, MessageId, ServerId};
use escape::escape;
use std::cmp;
use std::collections::HashMap;
use std::error::Error;
use std::fmt;
use std::fs::File;
use std::io::{BufRead, BufReader};
use std::io::Write;
use std::process::Command;

pub struct CommandContext {
	pub tokens: Vec<String>,
	pub selected: usize,

	pub session: Discord,
	pub gateway: Connection,
	pub state: State,

	pub guild: Option<ServerId>,
	pub channel: Option<ChannelId>,

	pub alias: HashMap<String, Vec<String>>,
	pub using: Option<Vec<String>>
}
impl ::std::fmt::Debug for CommandContext {
	fn fmt(&self, fmt: &mut ::std::fmt::Formatter) -> ::std::fmt::Result { write!(fmt, "context here") }
}
impl CommandContext {
	pub fn new(tokens: Vec<String>, selected: usize) -> Result<CommandContext, ::discord::Error> {
		let conn = ::connect(tokens[selected].as_str());
		if let Err(err) = conn {
			return Err(err);
		}
		let (session, gateway, state) = conn.unwrap();

		Ok(CommandContext {
			tokens: tokens,
			selected: selected,

			session: session,
			gateway: gateway,
			state: state,

			guild: None,
			channel: None,

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
				map.insert("silent".to_string(), vec!["to".to_string(), String::new()]);

				map
			},
			using: None
		})
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
pub fn execute(context: &mut CommandContext, terminal: bool, mut tokens: Vec<String>) -> CommandResult {
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
		($min:expr, $usage:expr) => {
			if tokens.len() < $min {
				fail!(concat!($usage, "\nYou supplied too few arguments"));
			}
		}
	}
	macro_rules! usage_max {
		($max:expr, $usage:expr) => {
			if tokens.len() > $max {
				fail!(concat!($usage, "\nYou supplied too many arguments"));
			}
		}
	}
	macro_rules! usage {
		($exact:expr, $usage:expr) => {
			usage_min!($exact, $usage);
			usage_max!($exact, $usage);
		}
	}
	macro_rules! usage_one {
		($usage:expr) => {
			if tokens.len() != 1 {
				fail!(concat!($usage, "\nYou did not supply 1 argument.\n\
								Did you mean to put quotes around the argument?"));
			}
		}
	}
	macro_rules! from_id {
		($type:expr, $funcid:ident, $funcname:ident, $ref:expr, $nameorid:expr) => {
			{
				let i = $nameorid.parse();
				let mut val;

				if i.is_err() {
					val = context.state.$funcname(context.guild, $nameorid.as_str())
				} else {
					val = context.state.$funcid($type(i.unwrap()));
					if val.is_none() {
						val = context.state.$funcname(context.guild, $nameorid.as_str())
					}
				}

				val
			}
		}
	}
	macro_rules! attempt {
		($result:expr, $message:expr) => {
			match $result {
				Err(err) => fail!(format!("{} (Details: {})", $message, err)),
				Ok(ok) => ok,
			}
		}
	}
	macro_rules! require {
		($option:expr, $message:expr) => {
			match $option {
				None => fail!($message),
				Some(some) => some,
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
		() => {
			require!(context.guild, "This command requires a guild to be selected")
		}
	}
	macro_rules! require_channel {
		() => {
			require!(context.channel, "This command requires a channel to be selected")
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
		($str:expr, $type:tt) => {
			{
				let num = $str.parse::<$type>();
				attempt!(num, "Not a number")
			}
		}
	}
	macro_rules! msg {
		($id:expr) => {
			{
				format!("Sent message with ID {}", $id)
			}
		}
	}
	macro_rules! max {
		($num:expr, $max:expr) => {
			{
				if $num > $max {
					fail!(format!("Too high. Max: {}", $max));
				}

				$num
			}
		}
	}

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
			usage_one!("echo <text>");
			success!(Some(tokens[0].clone()));
		},
		"help" => {
			usage_one!("help <command>");
			success!(Some(::help::about(tokens[0].as_str())))
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
					tokens.remove(0);
					if name == "alias" {
						fail!("lol nope");
					}

					if tokens.len() >= 2 && tokens[0] == "=" {
						tokens.remove(0);
						usage_min!(1, "alias <name> = <command...>");
					}
					context.alias.insert(name, tokens.to_vec());

					success!(None);
				},
			}
		},
		"exec" => {
			usage_min!(2, "exec <type> <value>");

			match tokens[0].as_str() {
				"shell" => {
					usage_max!(2, "exec shell <command>");

					let cmd = if cfg!(target_os = "windows") {
						Command::new("cmd")
							.arg("/c")
							.arg(tokens[1].clone())
							.status()
					} else {
						Command::new("sh").arg("-c").arg(tokens[1].clone()).status()

					};
					if cmd.is_err() {
						fail!(couldnt!("execute command"));
					}
					success!(Some(format!(
								"{}Process exited with status {}{}",
								if terminal { *COLOR_BLACK } else { "" },
								cmd.unwrap().code().unwrap_or(1),
								if terminal { *COLOR_RESET } else { "" },
							)));
				},
				"file" => {
					usage_max!(2, "exec file <file>");
					let result = execute_file(context, terminal, tokens[1].clone());
					let result = attempt!(result, couldnt!("run commands file"));

					success!(Some(result))
				},
				"lua" => {
					let mut lua = new_lua(context, terminal);

					let file = attempt!(File::open(tokens[1].clone()), couldnt!("open file"));
					if let Err(err) = lua.execute_from_reader::<(), _>(file) {
						fail!(format!("Error trying to execute: {:?}", err));
					}
					success!(None);
				},
				"lua-inline" => {
					usage_max!(2, "exec lua-inline <text>");
					let mut lua = new_lua(context, terminal);

					if let Err(err) = lua.execute::<()>(tokens[1].clone().as_str()) {
						fail!(format!("Error trying to execute: {:?}", err));
					}
					success!(None);
				},
				_ => fail!(unknown!("type (shell/file/lua/lua-inline available)")),
			}
		},
		"use" => {
			usage_min!(1, "use <command...>");

			context.using = Some(tokens);
			success!(Some(
				"Use mode enabled.\nSend an empty command to disable."
					.to_string()
			));
		},
		"to" => {
			usage_min!(2, "to <file> <command...>");

			let file = tokens[0].clone();
			tokens.remove(0);

			if tokens[0] == "from" {
				tokens.remove(0);
				usage_min!(1, "to <file> from <command...>");
			}

			let mut result = execute(context, false, tokens);

			if file.is_empty() {
				if result.success {
					result.text = None;
				}
			} else {
				let file = File::create(file);
				let mut file = attempt!(file, couldnt!("open file"));

				if let Some(text) = result.text.clone() {
					let write = file.write_all(text.as_bytes());
					attempt!(write, couldnt!("write to file"));
					if result.success {
						result.text = None;
					}
				}
			}

			result
		},
		"accounts" => {
			usage_max!(1, "accounts [index]");

			match tokens.get(0) {
				None => {
					let mut output = String::new();
					let mut first = true;

					for (i, token) in context.tokens.iter().enumerate() {
						if first {
							first = false;
						} else {
							output.push('\n');
						}

						output.push_str(format!("{}. {}", i, token).as_str());
					}

					success!(Some(output));
				},
				Some(index) => {
					let index = parse!(index, usize);
					let token = match context.tokens.get(index) {
						None => fail!("Out of bounds"),
						Some(token) => token,
					};

					context.selected = index;

					let conn = ::connect(token);
					let (session, gateway, state) = attempt!(conn, "Could not connect to gateway");

					// context.gateway.shutdown();
					//
					// The borrow checker hates me.

					context.session = session;
					context.gateway = gateway;
					context.state = state;

					success!(None);
				},
			}
		},
		"exit" => {
			usage_max!(0, "exit");
			CommandResult {
				exit: true,
				..Default::default()
			}
		},
		"guild" => {
			usage_max!(1, "guild [id/name]");

			if tokens.is_empty() {
				context.guild = None;
				context.channel = None;
				success!(None);
			}
			let guild = from_id!(
				ServerId,
				find_guild,
				find_guild_by_name,
				&mut guild,
				tokens[0]
			);

			let guild = unwrap_cache!(guild);
			context.guild = Some(guild.id);
			context.channel = Some(guild.id.main());

			success!(Some(pretty_json!({
						"id":       guild.id.to_string().as_str(),
						"name":     guild.name.as_str(),
						"owner_id": guild.owner_id.to_string().as_str(),
					})));
		},
		"channel" => {
			usage_max!(1, "channel [id/name]");

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
			usage_max!(0, "guilds");
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
			usage_max!(0, "channels");
			let guild = require_guild!();
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
			usage!(3, "msg <type> <\"send\"/existing id> <text>");
			let channel = require_channel!();

			let kind = match tokens[0].clone().as_str() {
				"normal" => 0,
				"tts" => 1,
				"embed" => 2,
				_ => fail!(unknown!("type (normal/tts/embed available)")),
			};
			let edit = match tokens[1].clone().as_str() {
				"send" => None,
				id => Some(parse!(id, u64)),
			};

			let text = tokens[2].clone();
			let mut text = text.as_str();

			match kind {
				0 | 1 => {
					if let Some(edit) = edit {
						if kind == 1 {
							fail!("Can't edit TTS");
						}
						max!(text.len() as u16, LIMIT_MSG);
						let msg = context.session.edit_message(channel, MessageId(edit), text);
						let msg = attempt!(msg, couldnt!("send message"));
						success!(Some(msg!(msg.id)));
					} else {
						let mut output = String::new();

						let mut first = true;
						while !text.is_empty() {
							if first {
								first = false;
							} else {
								output.push('\n');
							}
							let amount = cmp::min(text.len(), LIMIT_MSG as usize);
							let value = &text[..amount];
							text = &text[amount..];
							let msg = context.session.send_message(channel, value, "", kind == 1);
							let msg = attempt!(msg, couldnt!("send message"));
							output.push_str(msg!(msg.id).as_str());
						}

						success!(Some(output));
					}
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
		},
		"log" => {
			usage_max!(1, "log [n=10]");
			let channel = require_channel!();

			let limit = match tokens.get(0) {
				Some(num) => Some(max!(parse!(num, u16), LIMIT) as u64), // Ugh. discord-rs uses u64 even though even u16 is more than enough
				None => Some(10),
			};

			let messages = context.session.get_messages(
				channel,
				GetMessages::MostRecent,
				limit
			);
			let messages = attempt!(messages, couldnt!("get messages"));

			let mut output = String::new();
			let mut first = true;
			for msg in messages.iter().rev() {
				if first {
					first = false;
				} else {
					output.push('\n');
				}

				if terminal {
					output.push_str(*COLOR_CYAN);
				}

				let name = msg.author.name.as_str();
				let discrim = msg.author.discriminator.to_string();
				let discrim = discrim.as_str();

				output.reserve(name.len() + 1 + discrim.len());
				output.push_str(name);
				output.push('#');
				output.push_str(discrim);
				if terminal {
					output.push_str(*COLOR_RESET);
				}
				output.push_str(": ");
				output.push_str(msg.content.as_str())
			}

			success!(Some(output));
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

pub fn execute_file(context: &mut CommandContext, terminal: bool, file: String) -> Result<String, Box<Error>> {
	let file = File::open(file)?;
	let bufreader = BufReader::new(file);

	let pointer = ::raw::pointer(context, terminal);

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
		let tokens = ::tokenizer::tokens(|| if first {
			first = false;
			Ok(line.clone())
		} else {
			Err(ErrUnclosed)
		});
		let tokens = tokens?;
		let result = execute(context, terminal, tokens);

		if result.empty {
			continue;
		}
		if result.exit {
			results.push_str("Can't exit from a commands file");
			continue;
		}

		results.push_str(pointer.clone().as_str());
		if terminal {
			results.push_str(*COLOR_ITALIC);
		}
		results.push_str(line.as_str());
		if terminal {
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
				.map(|value| {
					let value0 = lua_to_string(value.0.clone());
					let value1 = lua_to_string(value.1.clone());
					let mut string = String::with_capacity(value0.len() + 2 + value1.len());
					string.push_str(value0.as_str());
					string.push_str(": ");
					string.push_str(value1.as_str());

					string
				})
				.collect::<Vec<_>>()
				.join(", ")
		},
		_ => String::new(),
	}

}
pub fn new_lua(context: &mut CommandContext, terminal: bool) -> Lua {
	let mut lua = Lua::new();
	lua.openlibs();

	// Example: `cmd({"echo", "Hello World"})`
	// crashes on incorrect type; see https://github.com/tomaka/hlua/issues/149
	lua.set(
		"cmd",
		hlua::function1::<_, String, Vec<AnyLuaValue>>(move |args| {
			let args = args.iter()
				.map(|value| lua_to_string(value.clone()))
				.collect();
			execute(context, terminal, args).text.unwrap_or_default()
		})
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
