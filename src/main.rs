extern crate discord;

macro_rules! stderr {
	($fmt:expr, $($arg:tt)*) => { writeln!(::std::io::stderr(), concat!($fmt, "\n"), $($arg)*).unwrap(); }
}
macro_rules! flush {
	() => { ::std::io::stdout().flush().unwrap(); }
}

mod options;
mod tui;
mod tokenizer;
mod command;

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

	//let session = discord::Discord::from_user_token(options.tokens[options.token].as_str()).unwrap();

	println!("{:?}", options.tokens);
	tui::tui();
}
