fn tokens(input: &str) -> Vec<String> {
	let mut tokens = vec![];
	let mut buffer = String::new();

	let mut escaped = true;
	let mut in_quote = '\0';

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
	if !buffer.is_empty() {
		tokens.push(buffer);
	}
	tokens
}

#[test]
fn test_tokens() {
	assert_eq!(tokens("hello `world \\` lol` r\\ i\\ p"), vec!["hello", "world ` lol", "r i p"])
}
