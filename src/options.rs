use super::clap::{App, Arg};

#[derive(Debug)]
pub struct Options {
	tokens: Vec<String>,
}

pub fn get_options() -> Options {
	let args =
		App::new("discord_console")
		.version(super::VERSION)
		.about("Use discord in a new way")
		.author("LEGOlord208")
		.arg(Arg::with_name("token")
			.long("token")
			.short("t")
			.help("Specify Discord token")
			.multiple(true)
			.takes_value(true))
		.get_matches();

	let tokens = args.values_of("token");
	Options{
		tokens: if tokens.is_some() { tokens.unwrap().map(|s| s.to_string()).collect() } else { vec!() },
	}
}
