macro_rules! success {
	($val:expr) => { return ($val, true); }
}
macro_rules! fail {
	($val:expr) => { return (Some($val.to_string()), false); }
}
macro_rules! usage_min {
	($tokens:expr, $min:expr, $usage:expr) => {
		if $tokens.len() < $min {
			fail!(concat!($usage, "\nYou supplied too few arguments."));
		}
	}
}
macro_rules! usage_max {
	($tokens:expr, $max:expr, $usage:expr) => {
		if $tokens.len() > $max {
			fail!(concat!($usage, "\nYou supplied too many arguments."));
		}
	}
}
macro_rules! usage {
	($tokens:expr, $exact:expr, $usage:expr) => {
		usage_min!($tokens, $exact, $usage);
		usage_max!($tokens, $exact, $usage);
	}
}
macro_rules! usage_one {
	($tokens:expr, $usage:expr) => {
		if $tokens.len() != 1 {
			fail!(concat!($usage, "\nYou did not supply 1 argument.\n \
							Did you mean to put quotes around the argument?"));
		}
	}
}

/*
pub struct CommandContext {
	guild:   Option<String>,
	channel: Option<String>,
}
*/

pub fn execute<'a>(mut tokens: Vec<String>/*, context: CommandContext*/) -> (Option<String>, bool) {
	if tokens.len() < 1 {
		success!(None);
	}
	let command = tokens.remove(0);
	let command = command.as_str();

	match command {
		"echo" => {
			usage_one!(tokens, "echo <text>");
			success!(Some(tokens[0].clone()));
		},
		_ => {
			fail!("Unknown command!");
		},
	}
}

#[cfg(test)]
mod test {
	#[test]
	fn test_execute() {
		assert_eq!(super::execute(vec!["echo".to_string(), "Hello World".to_string()]), (Some("Hello World".to_string()), true))
	}
}
