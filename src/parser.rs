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

use clipboard::{ClipboardContext, ClipboardProvider};

pub fn parse<GET, ERR>(mut input: GET) -> Result<Vec<String>, ERR>
where
	GET: FnMut() -> Result<String, ERR>,
{

	let mut tokens = Vec::new();
	let mut buffer = String::new();

	let mut escaped = false;
	let mut in_quote = '\0';

	let mut first = true;
	loop {
		if first {
			first = false;
		} else {
			buffer.push(' ');
		}
		let input = input()?;
		for c in input.trim().chars() {
			if c == '\\' && !escaped {
				escaped = true;
				continue;
			}
			if escaped {
				escaped = false;

				match c {
					'n' => buffer.push('\n'),
					'c' => {
						// Unsure where and how to save the variable 'clip' to
						// be the most efficient.
						// Storing it globally seems to make it complain A LOT.
						let clip = ClipboardProvider::new();
						if clip.is_ok() {
							let mut clip: ClipboardContext = clip.unwrap();
							if let Ok(clip) = clip.get_contents() {
								buffer.push_str(&clip);
							}
						}
					},
					_ => buffer.push(c),
				}
				continue;
			}
			match c {
				'"' | '`' => {
					if in_quote == '\0' {
						in_quote = c;
					} else if in_quote == c {
						in_quote = '\0';
					} else {
						buffer.push(c);
					}
				},
				' ' => {
					if in_quote == '\0' {
						tokens.push(buffer);
						buffer = String::new();
					} else {
						buffer.push(c);
					}
				},
				_ => buffer.push(c),
			}
		}
		if in_quote == '\0' && !escaped {
			if !buffer.is_empty() {
				tokens.push(buffer);
			}
			break;
		}
	}
	Ok(tokens)
}

#[cfg(test)]
#[test]
fn test() {
	macro_rules! test {
		($str:expr, $vec:expr) => {
			assert_eq!(parse::<_, ()>(|| Ok($str.to_string())).unwrap(), $vec);
		}
	}
	// General test.
	test!(
		"hello `world \\` lol` l\\ o\\ l",
		vec!["hello", "world ` lol", "l o l"]
	);

	// More escaping.
	test!("hello\" world\\\\\"", vec!["hello world\\"]);

	// Calls again and if quote is unclosed
	test!("hello\"", vec!["hello hello"]);
}
