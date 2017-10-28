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

use std::env;

lazy_static! {
	pub static ref NOCOLOR: bool = {
		env::var("TERM").unwrap_or_default() == "dumb"
	};
	pub static ref COLOR_BLACK: &'static str = {
		if *NOCOLOR { "" } else { "\x1B[0;30m" }
	};
	pub static ref COLOR_RED: &'static str = {
		if *NOCOLOR { "" } else { "\x1B[1;31m" }
	};
	pub static ref COLOR_YELLOW: &'static str = {
		if *NOCOLOR { "" } else { "\x1B[0;33m" }
	};
	pub static ref COLOR_CYAN: &'static str = {
		if *NOCOLOR { "" } else { "\x1B[0;36m" }
	};

	pub static ref COLOR_ITALIC: &'static str = {
		if *NOCOLOR { "" } else { "\x1B[3m" }
	};
	pub static ref COLOR_RESET: &'static str = {
		if *NOCOLOR { "" } else { "\x1B[0m" }
	};
}
