fn tokens<'a, GET>(input: GET) -> Result<Vec<String>, &'a str>
	where GET: Fn() -> Result<&'a str, &'a str> {

	let mut tokens = Vec::new();
	let mut buffer = String::new();

	let mut escaped = true;
	let mut in_quote = '\0';

	loop {
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
	#[test]
	fn test_tokens() {
		// General test.
		assert_eq!(super::tokens(|| Ok("hello `world \\` lol` r\\ i\\ p")).unwrap(), vec!["hello", "world ` lol", "r i p"]);

		// More escaping.
		assert_eq!(super::tokens(|| Ok("hello\" world\\\\\"")).unwrap(), vec!["hello world\\"]);

		// Calls again if quote is not closed.
		assert_eq!(super::tokens(|| Ok("hello\"")).unwrap(), vec!["hellohello"]);
	}
}
