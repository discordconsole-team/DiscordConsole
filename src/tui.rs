extern crate cursive;

use self::cursive::Cursive;
use self::cursive::event::Key;
use self::cursive::menu::MenuTree;
use self::cursive::view::{Offset, Position};
use self::cursive::views::{Dialog, EditView, LinearLayout, TextView};

use std::sync::{Arc, Mutex};
use tui::cursive::traits::Identifiable;

pub fn tui(context: ::command::CommandContext) {
	let mut screen = Cursive::new();
	screen.add_global_callback('q', Cursive::quit);

	let mut guilds = MenuTree::new();
	let servers = if let Some(settings) = context.state.settings() {
		::sort::sort_guilds(&settings, context.state.servers().to_vec())
	} else {
		context.state.servers().to_vec()
	};
	for server in servers {
		guilds.add_leaf(
			server.name.clone(), move |s| {
				s.screen_mut()
					.add_layer(
						Dialog::around(LinearLayout::vertical().child(TextView::new(format!("Name: {}", server.name))))
							.title("Guild info")
							.dismiss_button("Close")
					);
			}
		);
	}

	let context = Arc::new(Mutex::new(context));

	screen
		.menubar()
		.add_subtree(
			"General",
			MenuTree::new()
				.leaf(
					"Run command", move |s| {
						// if let Some(dialog) = s.find_id("cmd").unwrap() {
						if s.find_id::<Dialog>("cmd").is_some() {
							s.pop_layer();
							return;
						}

						let ctx = context.clone();

						s.screen_mut()
							.add_layer_at(
								Position {
									x: Offset::Absolute(5),
									y: Offset::Absolute(5)
								},
								Dialog::around(
									EditView::new().on_submit(
										move |s, string| {
											let mut first = true;
											let tokens = ::tokenizer::tokens::<_, ()>(
												|| if first {
													first = false;
													Ok(string.to_string())
												} else {
													Err(())
												}
											);
											if tokens.is_err() {
												s.add_layer(Dialog::info("Unclosed quote or trailing \\."));
												return;
											}

											command(s, &mut ctx.lock().unwrap(), tokens.unwrap());
										}
									)
								)
										.title("Run command")
										.with_id("cmd")
							);
					}
				)
				.leaf("Exit", Cursive::quit)
		)
		.add_subtree("Guilds", guilds);

	screen.add_layer(Dialog::text("Press <esc> to access the menu, and <q> to quit"));

	screen.add_global_callback(Key::Esc, |s| s.select_menubar());
	screen.set_autohide_menu(false);
	screen.run();
}

/*
fn command_field(context: Arc<Mutex<::command::CommandContext>>, key: &str, val: &str, tokens: Vec<&str>) -> Button {
	let mut string = String::with_capacity(key.len() + 2 + val.len());
	string.push_str(key);
	string.push_str(": ");
	string.push_str(val);

	let mut tokens = Arc::new(Mutex::new(tokens.iter().map(|string| string.to_string()).collect()));
	let val = val.to_string();
	Button::new(
		string, move |s| {
			let ctx = context.clone();
			s.add_layer(
				EditView::new()
					.content(val.clone())
					.on_submit_mut(
						move |s, string| {
							let tokens: &mut Vec<String> = &mut tokens.lock().unwrap();
							tokens.push(string.to_string());
							command(s, &mut ctx.lock().unwrap(), tokens.clone());
						}
					)
			);
		}
	)
}
*/

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
			None => {
				if result.success {
					text.push_str("Successfully executed command.");
				} else {
					text.push_str("Failed to execute command.");
				}
			},
		}
		s.add_layer(Dialog::info(text.as_str()))
	}
	result
}
