macro_rules! success {
	($val:expr) => {
		return CommandResult{
			text:    $val,
			success: true,
			exit:    false,
		}
	}
}
macro_rules! fail {
	($val:expr) => {
		return CommandResult{
			text:    Some($val.to_string()),
			success: false,
			exit:    false,
		}
	}
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
			fail!(concat!($usage, "\nYou did not supply 1 argument.\n\
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
pub struct CommandResult {
	pub text:    Option<String>,
	pub success: bool,
	pub exit:    bool,
}

pub fn execute<'a>(mut tokens: Vec<String>/*, context: CommandContext*/) -> CommandResult {
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
		"exit" => {
			usage_max!(tokens, 0, "exit");
			CommandResult{
				text:    None,
				success: true,
				exit:    true,
			}
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
		assert_eq!(super::execute(vec!["echo".to_string(), "Hello World".to_string()]).text.unwrap(), "Hello World".to_string())
	}
}
