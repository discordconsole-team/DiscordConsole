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
pub fn escape(token: &str) -> String {
	let mut escaped = String::with_capacity(token.len());
	let mut found = false;
	let mut empty = true;

	for c in token.chars() {
		empty = false;
		match c {
			'\\' => escaped.push_str("\\\\"),
			'"' => escaped.push_str("\\\""),
			' ' => {
				found = true;
				escaped.push(' ');
			},
			_ => escaped.push(c),
		}
	}

	if found || empty {
		escaped.insert(0, '\"');
		escaped.push('\"');
	}

	escaped
}

#[cfg(test)]
mod test {
	#[test]
	fn test_escape() {
		assert_eq!(super::escape("hello world"), "\"hello world\"");
		assert_eq!(super::escape("=)"), "=)");
		assert_eq!(super::escape("\\"), "\\\\");
	}
}
