extern crate cursive;
extern crate clap;

mod options;

use cursive::Cursive;
use cursive::event::Key;
use cursive::views::{Dialog, EditView};
use cursive::menu::MenuTree;

const VERSION: &str = "0.1";

fn main() {
	let options = options::get_options();

	println!("{:?}", options);

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
