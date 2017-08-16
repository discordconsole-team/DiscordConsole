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
use discord::model::{ChannelId, ChannelType, Game, LiveServer, MessageId, OnlineStatus, ServerId, UserId};
use escape::escape;
use help;
use std::cmp;
use std::collections::HashMap;
use std::error::Error;
use std::fmt;
use std::fs::File;
use std::io::{BufRead, BufReader, Write};
use std::process::Command;

const UPDATE_STATUS_HELP: &str = "Hey there buddy! You seem confused over how you should set status!\n\
					 Don't worry, it's not that hard once you get used to it.\n\n\

					 The first most important thing is: There is barely any order.\n\
					 Saying `update status stream \"a game\" \"url\"` is the same as \
						`update status \"a game\" stream \"url\"`\n\
					 There is some order though. URL has to come after game. And \"stream\" has to come before URL.\n\n\

					 How did you trigger this message? You tried to set an unknown value, \
						but both game and url is already set\n\
					 - or you're not streaming.\n\n\

					 There are also \"known\" values. These are: online, idle, donotdisturb, invisible, offline, stream.\n\
					 Every one of these values represent your status.\n\n\

					 What does this all mean in practise? Let's see!\n\
					 `update status \"a game\"` sets the user to online and with that playing status.\n\
					 `update status idle` sets the user to idle and no gaming status.\n\
					 `update status \"a game\" stream \"url\"` sets the streaming status to \"a game\" and the \
						URL to \"url\".\n\
					 `update status \"a game\" \"url\" stream` displays this message, because URL wasn't \
						expected since stream wasn't set.\n\
					 `update status online online` sets the playing status to \"online\", since we already set status.\n\
					 `update status please help` displays this message, since we're not streaming, but a second \
						unknown value was given.";

pub struct CommandContext {
	pub tokens: Vec<String>,
	pub selected: usize,

	pub session: Discord,
	pub gateway: Connection,
	pub state: State,

	pub guild: Option<ServerId>,
	pub channel: Option<ChannelId>,

	pub alias: HashMap<String, Vec<String>>,
	pub using: Option<Vec<String>>,

	pub ptr0: String,
	pub ptr1: String,
	pub ptr2: String
}
impl ::std::fmt::Debug for CommandContext {
	fn fmt(&self, fmt: &mut ::std::fmt::Formatter) -> ::std::fmt::Result { write!(fmt, "context here") }
}
impl CommandContext {
	pub fn new(tokens: Vec<String>, selected: usize) -> Result<CommandContext, ::discord::Error> {
		let conn = ::connect(&tokens[selected]);
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
			using: None,

			ptr0: "> ".to_string(),
			ptr1: "%c> ".to_string(),
			ptr2: "%g (%c)> ".to_string()
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

// Unsure if I really should split the function up. It shall be thought about.
// Shut clippy up about my macros... for now at least
#[cfg_attr(feature = "cargo-clippy", allow(needless_return))]
#[cfg_attr(feature = "cargo-clippy", allow(cyclomatic_complexity))]
pub fn execute(context: &mut CommandContext, terminal: bool, mut args: Vec<String>) -> CommandResult {
	macro_rules! success {
		($val:expr) => {
			return CommandResult{
				text:    $val,
				..Default::default()
			};
		}
	}
	macro_rules! fail {
		($val:expr) => {
			return CommandResult{
				text:    Some($val.to_string()),
				success: false,
				..Default::default()
			};
		}
	}
	macro_rules! usage_min {
		($min:expr, $command:expr) => {
			if args.len() < $min {
				fail!(format!("{}\n\nYou supplied too few arguments", help::about($command)));
			}
		}
	}
	macro_rules! usage_max {
		($max:expr, $command:expr) => {
			if args.len() > $max {
				fail!(format!("{}\n\nYou supplied too many arguments", help::about($command)));
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
		($command:expr) => {
			if args.len() != 1 {
				fail!(format!("{}\n\nYou did not supply 1 argument.\n\
								Did you mean to put quotes around the argument?", help::about($command)));
			}
		}
	}
	macro_rules! from_id {
		(kind: $type:expr, funcid: $funcid:ident, funcname: $funcname:ident, name_or_id: $nameorid:expr) => {
			{
				let i = $nameorid.parse();
				let mut val;

				if i.is_err() {
					val = context.state.$funcname(context.guild, &$nameorid)
				} else {
					val = context.state.$funcid($type(i.unwrap()));
					if val.is_none() {
						val = context.state.$funcname(context.guild, &$nameorid)
					}
				}

				val
			}
		}
	}
	macro_rules! attempt {
		($result:expr, $message:expr) => {
			match $result {
				Err(err) => fail!(format!("{} (Details: {:?})", $message, err)),
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
		($str:expr, $type:ty) => {
			{
				let num = $str.parse::<$type>();
				attempt!(num, "Not a number")
			}
		}
	}
	macro_rules! parse_user {
		($str:expr) => {
			{
				if $str == "@me" {
					unimplemented!();
				} else {
					UserId(parse!($str, u64))
				}
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
	macro_rules! min {
		($num:expr, $min:expr) => {
			{
				if $num < $min {
					fail!(format!("Too low. Min: {}", $min));
				}

				$num
			}
		}
	}
	macro_rules! require_bot {
		() => {
			if !context.tokens[context.selected].starts_with("Bot ") {
				fail!("Only bots can use this endpoint");
			}
		}
	}

	if args.len() < 1 {
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

		using.append(&mut args);
		args = using;
	}
	if let Some(atokens) = context.alias.get(&args[0]) {
		let mut atokens = atokens.clone();

		args.remove(0);
		atokens.append(&mut args);
		args = atokens;
	}
	let command = args.remove(0);

	match &*command {
		"echo" => {
			usage_one!("echo");
			success!(Some(args.remove(0)));
		},
		"help" => {
			usage_one!("help");
			success!(Some(::help::about(&args[0])))
		},
		"alias" => {
			match args.len() {
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
						output.push_str(&escape(key));
						output.push_str(" = ");
						output.push_str(&val.iter()
							.map(|item| escape(item))
							.collect::<Vec<String>>()
							.join(" "));
					}

					success!(if output.is_empty() {
						None
					} else {
						Some(output)
					});
				},
				1 => {
					context.alias.remove(&args[0]);
					success!(None);
				},
				_ => {
					let name = args.remove(0);
					if name == "alias" {
						fail!("lol nope");
					}

					if args.len() >= 2 && args[0] == "=" {
						args.remove(0);
						usage_min!(1, "alias");
					}
					context.alias.insert(name, args);

					success!(None);
				},
			}
		},
		"exec" => {
			usage_min!(2, "exec");

			match &*args[0] {
				"shell" => {
					usage_max!(2, "exec");

					let cmd = if cfg!(target_os = "windows") {
						Command::new("cmd").arg("/c").arg(&args[1]).status()
					} else {
						Command::new("sh").arg("-c").arg(&args[1]).status()
					};
					attempt!(cmd, couldnt!("execute command"));
					success!(Some(format!(
						"{}Process exited with status {}{}",
						if terminal { *COLOR_BLACK } else { "" },
						cmd.unwrap().code().unwrap_or(1),
						if terminal { *COLOR_RESET } else { "" },
					)));
				},
				"file" => {
					usage_max!(2, "exec");
					let result = execute_file(context, terminal, &args[1]);
					let result = attempt!(result, couldnt!("run commands file"));

					success!(Some(result))
				},
				"lua" => {
					let mut lua = new_lua(context, terminal);

					let file = attempt!(File::open(&args[1]), couldnt!("open file"));
					if let Err(err) = lua.execute_from_reader::<(), _>(file) {
						fail!(format!("Error trying to execute: {:?}", err));
					}
					success!(None);

					// TODO: Arguments
				},
				"lua-inline" => {
					let mut lua = new_lua(context, terminal);

					if let Err(err) = lua.execute::<()>(&args[1]) {
						fail!(format!("Error trying to execute: {:?}", err));
					}
					success!(None);

					// TODO: Arguments
				},
				_ => fail!(unknown!("type (shell/file/lua/lua-inline available)")),
			}
		},
		"use" => {
			usage_min!(1, "use");

			context.using = Some(args);
			success!(Some(
				"Use mode enabled.\nSend an empty command to disable."
					.to_string()
			));
		},
		"to" => {
			usage_min!(2, "to");

			let file = args.remove(0);

			if args[0] == "from" {
				args.remove(0);
				usage_min!(1, "to <file> from <command...>");
			}

			let mut result = execute(context, false, args);

			if file.is_empty() {
				if result.success {
					result.text = None;
				}
			} else {
				let file = File::create(file);
				let mut file = attempt!(file, couldnt!("open file"));

				if let Some(ref text) = result.text {
					let write = file.write_all(text.as_bytes());
					attempt!(write, couldnt!("write to file"));
				}
				if result.success {
					result.text = None;
				}
			}

			result
		},
		"accounts" => {
			usage_max!(1, "accounts");

			match args.get(0) {
				None => {
					let mut output = String::new();
					let mut first = true;

					for (i, token) in context.tokens.iter().enumerate() {
						if first {
							first = false;
						} else {
							output.push('\n');
						}

						output.push_str(&format!("{}. {}", i, token));
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
			usage_max!(1, "guild");

			if args.is_empty() {
				context.guild = None;
				context.channel = None;
				success!(None);
			}
			let guild =
				from_id!(
				kind: ServerId,
				funcid: find_server,
				funcname: find_guild_by_name,
				name_or_id: args[0]
			);

			let guild = unwrap_cache!(guild);
			context.guild = Some(guild.id);
			let mut main = None;
			for channel in &guild.channels {
				if channel.kind == ChannelType::Text && channel.position == 0 {
					main = Some(channel);
					break;
				}
			}
			if let Some(main) = main {
				context.channel = Some(main.id);
			}

			success!(Some(pretty_json!({
				"id":       &guild.id.to_string(),
				"name":     &guild.name,
				"owner_id": &guild.owner_id.to_string(),
			})));
		},
		"channel" => {
			usage_max!(1, "channel");

			if args.is_empty() {
				if let Some(guild) = context.guild {
					context.channel = Some(guild.main());
				} else {
					context.channel = None;
				}
				success!(None);
			}
			let channel = from_id!(
				kind: ChannelId,
				funcid: find_channel,
				funcname: find_channel_by_name,
				name_or_id: args[0]
			);
			let channel = unwrap_cache!(channel);

			match channel {
				ChannelRef::Private(channel) => {
					context.guild = None;
					context.channel = Some(channel.id);

					success!(Some(pretty_json!({
						"id":       &channel.id.to_string(),
						"recipient": {
							"id":   &channel.recipient.id.to_string(),
							"name": &channel.recipient.name
						}
					})));
				},
				ChannelRef::Group(channel) => {
					context.guild = None;
					context.channel = Some(channel.channel_id);

					success!(Some(pretty_json!({
						"id":       &channel.channel_id.to_string(),
						"name":     &channel.name.clone().unwrap_or_default()
					})));
				},
				ChannelRef::Public(guild, channel) => {
					context.guild = Some(guild.id);
					context.channel = Some(channel.id);

					success!(Some(pretty_json!({
						"id":       &channel.id.to_string(),
						"name":     &channel.name,
						"guild": {
							"id":   &guild.id.to_string(),
							"name": &guild.name
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
				value.push_str(&guild.id.to_string());
				value.push(' ');
				value.push_str(&guild.name);
			}

			success!(Some(value));
		},
		"channels" => {
			usage_max!(0, "channels");
			let guild = require_guild!();
			let guild = unwrap_cache!(context.state.find_server(guild));

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
					value.push_str(&channel.id.to_string());
					value.push(' ');
					value.push_str(channel.kind.name());
					value.push(' ');
					value.push_str(&channel.name);
				}
			}

			success!(Some(value));
		},
		"msg" => {
			usage!(3, "msg");
			let channel = require_channel!();

			let kind = match &*args[0] {
				"normal" => 0,
				"tts" => 1,
				"embed" => 2,
				_ => fail!(unknown!("type (normal/tts/embed available)")),
			};
			let edit = match &*args[1] {
				"send" => None,
				id => Some(parse!(id, u64)),
			};

			let mut text = &*args[2];

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
							output.push_str(&msg!(msg.id));
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
		"del" => {
			usage!(1, "del");
			let channel = require_channel!();

			let message_str = &args[0];
			let message;
			let which;

			if message_str == ".." {
				message = None;
				which = Some(GetMessages::MostRecent);
			} else if message_str.starts_with("..") {
				message = Some(MessageId(parse!(message_str[2..], u64)));
				which = Some(GetMessages::Before(message.unwrap()));
			} else if message_str.ends_with("..") {
				message = Some(MessageId(parse!(message_str[..message_str.len() - 2], u64)));
				which = Some(GetMessages::After(message.unwrap()));
			} else {
				message = Some(MessageId(parse!(message_str, u64)));
				which = None;
			}

			if let Some(which) = which {
				let messages = context.session.get_messages(
					channel,
					which,
					Some(LIMIT as u64)
				);
				let messages = attempt!(messages, couldnt!("get messages"));
				if !messages.is_empty() {
					let ids: Vec<_> = messages.iter().map(|msg| msg.id).collect();
					let delete = context.session.delete_messages(channel, &ids);
					attempt!(delete, couldnt!("delete messages"));
				}
			} else {
				let delete = context.session.delete_message(channel, message.unwrap());
				attempt!(delete, couldnt!("delete message"));
			}
			success!(None);
		},
		"log" => {
			usage_max!(1, "log");
			let channel = require_channel!();

			let mut limit = match args.get(0) {
				Some(num) => min!(parse!(num, i32), 1),
				None => 10,
			};

			let mut messages = Vec::new();
			let mut which = GetMessages::MostRecent;
			while limit > 0 {
				let new_messages = context.session.get_messages(
					channel,
					which,
					Some(cmp::min(limit as u64, LIMIT as u64))
				);
				let mut new_messages = attempt!(new_messages, couldnt!("get messages"));

				limit -= LIMIT as i32;

				if new_messages.is_empty() {
					break;
				}

				which = GetMessages::Before(new_messages.last().unwrap().id);
				messages.append(&mut new_messages);
			}

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
				output.push_str(
					&format!("(ID: {}) {}#{:04}: {}", msg.id, msg.author.name, msg.author.discriminator, msg.content)
				);
				if terminal {
					output.push_str(*COLOR_RESET);
				}
			}

			success!(Some(output));
		},
		"update" => {
			usage_min!(2, "update");

			match &*args[0] {
				"name" => {
					usage_max!(2, "update");
					require_bot!();

					let result = context.session.edit_profile(
						|profile| profile.username(&*args[1])
					);
					attempt!(result, couldnt!("update name"));
				},
				"status" => {
					let mut streaming = false;
					let mut status = None;
					let mut game = None;
					let mut url = None;

					for value in &args[1..] {
						if status.is_none() && !streaming && value == "stream" {
							streaming = true;
						} else {
							let parse_result = OnlineStatus::from_name(value);
							// value.parse::<OnlineStatus>()
							if status.is_none() && parse_result.is_some() {
								status = parse_result;
							} else if game.is_none() {
								game = Some(value.clone());
							} else if streaming && url.is_none() {
								url = Some(value.clone())
							} else {
								success!(Some(UPDATE_STATUS_HELP.to_string()));
							}
						}
					}

					if streaming && url.is_none() {
						success!(Some(UPDATE_STATUS_HELP.to_string()));
					}

					let game_status = game.map(|game| if streaming {
						Game::streaming(game, url.unwrap())
					} else {
						Game::playing(game)
					});
					let status = status.unwrap_or(OnlineStatus::Online);

					context.gateway.set_presence(
						game_status,
						status,
						status == OnlineStatus::Idle
					)
				},
				_ => fail!(unknown!("property (name, status available)")),
			}

			success!(None);
		},
		"user" => {
			usage!(3, "user");

			let guild = require_guild!();
			let user = parse_user!(args[0]);

			if args[1] == "nick" {
				let member = context.session.edit_member(
					guild,
					user,
					|builder| builder.nickname(&args[2])
				);
				attempt!(member, couldnt!("edit member"));
			} else {
				fail!(unknown!("property (nick available)"));
			}

			success!(None);
		},
		"set" => {
			usage!(2, "set");

			let val = args.remove(1);
			let key = &*args[0];

			match key {
				"ptr0" => context.ptr0 = val,
				"ptr1" => context.ptr1 = val,
				"ptr2" => context.ptr2 = val,
				_ => fail!(unknown!("property")),
			}

			success!(None);
		},
		"get" => {
			usage!(1, "get");

			success!(Some(match &*args[0] {
				"ptr0" => context.ptr0.clone(),
				"ptr1" => context.ptr1.clone(),
				"ptr2" => context.ptr2.clone(),
				_ => fail!(unknown!("property")),
			}));
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

pub fn execute_file(context: &mut CommandContext, terminal: bool, file: &str) -> Result<String, Box<Error>> {
	let file = File::open(file)?;
	let bufreader = BufReader::new(file);

	let prompt = ::raw::prompt(context, terminal);

	let mut results = String::new();
	let mut first = true;

	let mut lines = bufreader.lines();

	while let Some(line) = lines.next() {
		let mut line = line?;

		if first {
			first = false;
		} else {
			results.push('\n');
		}

		let mut first = true;
		let args = ::parser::parse::<_, Box<Error>>(|| if first {
			first = false;
			Ok(line.clone())
		} else {
			match lines.next() {
				Some(line2) => {
					let line2 = line2?;
					line.push(' ');
					line.push_str(&line2);
					Ok(line2)
				},
				None => Err(Box::new(ErrUnclosed)),
			}
		})?;
		let result = execute(context, terminal, args);

		if result.empty {
			continue;
		}
		if result.exit {
			results.push_str("Can't exit from a commands file");
			continue;
		}

		results.push_str(&prompt);
		if terminal {
			results.push_str(*COLOR_ITALIC);
		}
		results.push_str(&line);
		if terminal {
			results.push_str(*COLOR_RESET);
		}
		results.push('\n');
		results.push_str(&result.text.unwrap_or_default())
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
					string.push_str(&value0);
					string.push_str(": ");
					string.push_str(&value1);

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
	fn find_guild_by_name<'a>(&'a self, guild: Option<ServerId>, name: &str) -> Option<&'a LiveServer>;
	fn find_channel_by_name<'a>(&'a self, guild: Option<ServerId>, name: &'a str) -> Option<ChannelRef<'a>>;
}
impl MoreStateFunctionsSuperOriginalTraitNameExclusiveTM for State {
	fn find_guild_by_name<'a>(&'a self, _: Option<ServerId>, name: &str) -> Option<&'a LiveServer> {
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
