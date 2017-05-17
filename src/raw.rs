extern crate rustyline;

use std::io::Write;
use self::rustyline::Editor;
use self::rustyline::error::ReadlineError;

pub fn raw(mut context: ::command::CommandContext) {
	let mut rl = Editor::<()>::new();

	loop {
		let mut first = true;
		let mut command = String::new();

		let tokens = ::tokenizer::tokens(|| {
			let result = rl.readline(if first {
				first = false;
				"> "
			} else {
				""
			});

			match result {
				Ok(res) => {
					command.push_str(res.as_str());
					Ok(res)
				},
				Err(err) => {
					Err(err)
				},
			}
		});
		rl.add_history_entry(command.as_str());
		let tokens = match tokens {
			Ok(tokens) => tokens,
			Err(ReadlineError::Eof) | Err(ReadlineError::Interrupted) => {
				break;
			},
			Err(err) => {
				stderr!("Error reading line: {}", err);
				break;
			},
		};

		let result = ::command::execute(&mut context, tokens);
		if result.success {
			if let Some(text) = result.text {
				println!("{}", text.as_str());
			}
		} else {
			if let Some(text) = result.text {
				stderr!("{}", text.as_str());
			}
		}

		if result.exit {
			break;
		}
	}
}
