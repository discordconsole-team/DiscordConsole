pub fn escape(tokens: &[String]) -> String {
	let mut first = true;
	let mut output = String::new();

	for token in tokens {
		if first {
			first = false
		} else {
			output.push(' ');
		}

		let mut escaped = String::new();
		let mut found = false;

		for c in token.chars() {
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

		if found {
			escaped.insert(0, '\"');
			escaped.push('\"');
		}

		output.push_str(escaped.as_str());
	}
	return output;
}

#[cfg(test)]
mod test {
	#[test]
	fn test_escape() {
		assert_eq!(
			super::escape(&["hello world".to_string(), "=)".to_string()]),
			"\"hello world\" =)"
		);
	}
}
