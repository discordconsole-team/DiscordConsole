extern crate cursive;

use self::cursive::Cursive;
use self::cursive::event::Key;
use self::cursive::view::{Position, Offset};
use self::cursive::views::{Dialog, EditView};
use self::cursive::menu::MenuTree;

use ::std::sync::{Arc, Mutex};

pub fn tui(context: ::command::CommandContext) {
	let mut screen = Cursive::new();
	let context = Arc::new(Mutex::new(context));

	screen.add_global_callback('q', Cursive::quit);

	screen.menubar()
		.add_subtree("General",
			MenuTree::new()
			.leaf("Run command", move |s| {
				let ctx = context.clone();

				s.screen_mut().add_layer_at(
					Position{
						x: Offset::Absolute(5),
						y: Offset::Absolute(5),
					},
					Dialog::around(EditView::new()
						.on_submit(move |s, string| {
							let mut first = true;
							let tokens    = ::tokenizer::tokens::<_, ()>(|| {
								if first {
									first = false;
									Ok(string.to_string())
								} else {
									Err(())
								}
							});
							if tokens.is_err() {
								s.add_layer(Dialog::info("Unclosed quote or trailing \\."));
								return;
							}

							command(s, &mut ctx.lock().unwrap(), tokens.unwrap());
						})
					)
					.dismiss_button("Close")
					.title("Run command"));
			})
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

fn command(s: &mut Cursive, context: &mut ::command::CommandContext, tokens: Vec<String>) -> ::command::CommandResult {
	let result = ::command::execute(context, tokens);
	if result.exit {
		s.quit();
		return result;
	}
	if !result.empty {
		let mut text = String::new();
		if !result.success {
			text.push_str("Error: ");
		}
		match result.text.clone() {
			Some(string) => text.push_str(string.as_str()),
			None         => {
				if result.success {
					text.push_str("Successfully executed command.");
				} else {
					text.push_str("Failed to execute command.");
				}
			}
		}
		s.add_layer(Dialog::info(text.as_str()))
	}
	result
}
