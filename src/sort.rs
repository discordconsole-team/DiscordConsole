/* DiscordConsole is a software aiming to give you full control over
 * accounts, bots and webhooks!
 * Copyright (C) 2017  LEGOlord208 */
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

use discord::model::{LiveServer, PublicChannel, UserSettings};

pub fn sort_guilds(settings: &UserSettings, guilds: &mut Vec<LiveServer>) {
	let mut new_guilds = Vec::new();

	for guild_id in &settings.server_positions {
		for guild in guilds.iter() {
			if guild.id == *guild_id {
				new_guilds.push(guild.clone());
				break;
			}
		}
	}

	for guild in &new_guilds {
		let i = guilds.iter().position(|guild2| guild2.id == guild.id);
		if let Some(i) = i {
			guilds.remove(i);
		}
	}

	guilds.append(&mut new_guilds);
}

pub fn sort_channels(channels: &mut Vec<&PublicChannel>) { channels.sort_by_key(|c| c.position); }
