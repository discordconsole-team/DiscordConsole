extern crate cursive;

use cursive::Cursive;
use cursive::event::Key;
use cursive::views::Dialog;
use cursive::menu::MenuTree;

fn main() {
	let mut screen = Cursive::new();
	screen.add_global_callback('q', Cursive::quit);

	screen.menubar()
		.add_subtree("Guilds",
			 MenuTree::new()
				 .subtree("Example Server",
					MenuTree::new()
						.leaf("ID: 123456789", |_| {})
						.leaf("Name: Example Server", |s| s.add_layer(Dialog::info("This would let you edit the name...")))
				)
		);

	screen.add_layer(
		Dialog::text("Press <esc> to access the menu, and <q> to quit")
	);

	screen.add_global_callback(Key::Esc, |s| s.select_menubar());
	screen.set_autohide_menu(false);
	screen.run();
}
