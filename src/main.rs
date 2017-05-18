extern crate discord;

use discord::{Discord, State};
pub use std::io::Write;

#[macro_export]
macro_rules! stderr {
	($fmt:expr)              => { writeln!(::std::io::stderr(), concat!($fmt, "\n")).unwrap(); };
	($fmt:expr, $($arg:tt)*) => { writeln!(::std::io::stderr(), concat!($fmt, "\n"), $($arg)*).unwrap(); };
}
macro_rules! flush {
	() => { ::std::io::stdout().flush().unwrap(); }
}

mod options;
mod tokenizer;
mod command;
mod sort;
mod raw;
mod tui;

const VERSION: &str = "0.1";

fn main() {
	let mut options = options::get_options();

	for token in &mut options.tokens {
		*token = token.trim().to_string();
		let lower = token.to_lowercase();

		if lower.starts_with("bot ") {
			*token = "Bot ".to_string() + &token[4..];
		} else if lower.starts_with("user ") {
			*token = token[5..].to_string();
		}
	}

	let session = Discord::from_user_token(options.tokens[options.token].as_str()).unwrap();
	let (conn, ready) = match session.connect() {
		Ok((conn, ready)) => (conn, ready),
		Err(_) => {
			stderr!("Could not connect to websocket.");
			return;
		},
	};
	let state = State::new(ready);

	let context = command::CommandContext::new(session, conn, state);

	if options.notui {
		raw::raw(context);
	} else {
		tui::tui(context);
	}
}
