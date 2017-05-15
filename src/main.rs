extern crate cursive;
extern crate clap;
extern crate discord;

use cursive::Cursive;
use cursive::event::Key;
use cursive::views::{Dialog, EditView};
use cursive::menu::MenuTree;

macro_rules! stderr {
	($fmt:expr, $($arg:tt)*) => { writeln!(::std::io::stderr(), concat!($fmt, "\n"), $($arg)*).unwrap(); }
}
macro_rules! flush {
	() => { ::std::io::stdout().flush().unwrap(); }
}

mod options;

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

	let mut screen = Cursive::new();
	screen.add_global_callback('q', Cursive::quit);

	screen.menubar()
		.add_subtree("General",
			MenuTree::new()
			.leaf("Exit", Cursive::quit))
		.add_subtree("Guilds",
			MenuTree::new()
			.subtree("Example Server",
				MenuTree::new()
				.leaf("ID: 123456789", |_| {})
				.leaf("Name: Example Server", |s| {
					s.add_layer(
						Dialog::around(EditView::new()
							.on_submit(|s, _| {
								s.pop_layer();
							})
						)
						.title("Enter new server name:"));
				})));

	screen.add_layer(Dialog::text("Press <esc> to access the menu, and <q> to quit"));

	screen.add_global_callback(Key::Esc, |s| s.select_menubar());
	screen.set_autohide_menu(false);
	screen.run();
}
