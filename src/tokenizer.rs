pub fn tokens<'a, GET, ERR>(mut input: GET) -> Result<Vec<String>, ERR>
	where GET: FnMut() -> Result<String, ERR> {

	let mut tokens = Vec::new();
	let mut buffer = String::new();

	let mut escaped = false;
	let mut in_quote = '\0';

	let mut first = true;
	loop {
		if first {
			first = false;
		} else {
			buffer.push('\n');
		}
		let input = input()?;
		for c in input.chars() {
			if c == '\\' && !escaped {
				escaped = true;
				continue;
			}
			if escaped {
				escaped = false;

				buffer.push(c);
				continue;
			}
			match c {
				'"' | '`' => {
					if in_quote == '\0' {
						in_quote = c;
					} else if in_quote == c {
						in_quote = '\0';
					}
				}
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
			break
		}
	}
	Ok(tokens)
}

#[cfg(test)]
mod test {
	macro_rules! test {
		($str:expr, $vec:expr) => {
			assert_eq!(super::tokens::<_, ()>(|| Ok($str.to_string())).unwrap(), $vec);
		}
	}

	#[test]
	fn test_tokens() {
		// General test.
		test!("hello `world \\` lol` l\\ o\\ l", vec!["hello", "world ` lol", "l o l"]);

		// More escaping.
		test!("hello\" world\\\\\"", vec!["hello world\\"]);

		// Calls again if quote is not closed.
		test!("hello\"", vec!["hellohello"]);
	}
}
