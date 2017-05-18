/* DiscordConsole is a software aiming to give you full control over
 * accounts, bots and webhooks!
 * Copyright (C) 2017  LEGOlord208
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 * */
extern crate cursive;

use self::cursive::Cursive;
use self::cursive::event::Key;
use self::cursive::menu::MenuTree;
use self::cursive::view::{Offset, Position};
use self::cursive::views::{Button, Dialog, EditView, LinearLayout};
use std::cell::RefCell;
use std::rc::Rc;
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

	let context = Rc::new(RefCell::new(context));
	for server in servers {
		let context = context.clone();
		guilds.add_leaf(
			server.name.clone(), move |s| {
				let context = context.clone();
				s.screen_mut()
					.add_layer(
						Dialog::around(LinearLayout::vertical().child(command_field(context, "Name", server.name.as_str(), vec!["echo"])))
							.title("Guild info")
							.dismiss_button("Close")
					);
			}
		);
	}

	screen
		.menubar()
		.add_subtree(
			"General",
			MenuTree::new()
				.leaf(
					"Run command", move |s| {
						if s.find_id::<Dialog>("cmd").is_some() {
							s.pop_layer();
							return;
						}

						let context = context.clone();

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

											command(s, &mut context.borrow_mut(), tokens.unwrap());
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

fn command_field(context: Rc<RefCell<::command::CommandContext>>, key: &str, val: &str, tokens: Vec<&str>) -> Button {
	let mut string = String::with_capacity(key.len() + 2 + val.len());
	string.push_str(key);
	string.push_str(": ");
	string.push_str(val);

	let tokens: Rc<Vec<String>> = Rc::new(tokens.iter().map(|string| string.to_string()).collect());
	let val = val.to_string();
	Button::new(
		string, move |s| {
			let context = context.clone();
			let tokens = tokens.clone();
			s.add_layer(
				Dialog::around(
					EditView::new()
						.content(val.clone())
						.on_submit(
							move |s, string| {
								s.pop_layer();
								let mut tokens = (*tokens).clone();
								tokens.push(string.to_string());
								command(s, &mut context.borrow_mut(), tokens);
							}
						)
				)
						.title("Edit Field")
						.dismiss_button("Cancel")
			);
		}
	)
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
