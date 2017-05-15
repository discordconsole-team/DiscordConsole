extern crate cursive;

use self::cursive::Cursive;
use self::cursive::event::Key;
use self::cursive::views::{Dialog, EditView};
use self::cursive::menu::MenuTree;

pub fn tui() {
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
