pub fn escape(token: &str) -> String {
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
