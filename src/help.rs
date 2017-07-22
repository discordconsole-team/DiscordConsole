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
pub fn about(command: &str) -> String {
	match command {
		"echo" => {
			"echo <text>\n\
				Print out the text specified."
		},
		"help" => {
			"help <command>\n\
				The help command shows information about a command."
		},
		"alias" => {
			"alias [name] [=] [command...]\n\
				If 0 arguments, Print all aliases.\n\
				If 1 argument, Remove alias with [name].\n\
				If 2 or more arguments, Make so typing [name] would execute [command...].\n\
				Anything trailing <name> would append to <command>\n\n\
				For examples, see the built-in aliases"
		},
		"exec" => {
			"exec <type> <value>\n\
				Execute <type> command/code <value>"
		},
		"use" => {
			"use <command...>\n\
				Make every command in the future automatically prepend <command...>.\n\
				Send empty command to disable."
		},
		"to" => {
			"to <file> <command...>\n\
				Print the output of <command...> to <file> in case of success.\n\
				<file> may be \"\" (empty) to discard the output.\n\
				Also see the built-in silent alias."
		},
		"accounts" => {
			"accounts [index]\n\
				List all accounts, and switch to another account by specifying index."
		},
		"exit" => {
			"exit\n\
				Attempt to exit the console.\n\
				Won't always work - for example in scripts."
		},
		"guild" => {
			"guild [id/name]\n\
				Select the guild with id specified, or if it's a non-numerical value,\n\
				the first it finds with that name.\n\
				If guild unspecified, deselect guild."
		},
		"channel" => {
			"channel [id/name]\n\
				Select the channel with id specified, or if it's a non-numerical value,\n\
				the first it finds with that name in the current guild.\n\
				If no current guild, search globally.\n\
				If channel unspecified, select general channel."
		},
		"guilds" => {
			"guilds\n\
				Sort all guilds after the user's settings and print them."
		},
		"channels" => {
			"guilds\n\
				Sort all channels after their position and print them."
		},
		"msg" => {
			"msg <type> <\"send\"/existing id> <text>\n\
				Send a <type> message, or if 2nd argument is not \"send\",\n\
				edit the message with the specified id.\n\
				You might want to use the built-in say, embed, edit aliases instead."
		},
		"log" => {
			"log [n=10]\n\
				Print the last [n] messages (default 10)"
		},
		"update" => {
			"update <property> <value>\n\
				Update <property> on user to <value>.\n\
				Alternatively, there may be multiple values, depending on property.\n\n\

				Example: `update name test-bot`, `update status idle \"a game\"`"
		},
		"user" => {
			"user <user id> <property> <value>\n\
				Update <property> on <user id> to <value> within the current server.\n\

				Example: `user @me nick \"Hello World\"`"
		},
		_ => "No help available",
	}.to_string()
}
