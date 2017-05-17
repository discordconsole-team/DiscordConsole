use ::discord::model::{UserSettings, LiveServer};

pub fn sort_guilds(settings: &UserSettings, mut guilds: Vec<LiveServer>) -> Vec<LiveServer> {
	let mut new_guilds = Vec::new();

	for guild_id in &settings.server_positions {
		for guild in &guilds {
			if guild.id == *guild_id {
				new_guilds.push(guild.clone());
				break;
			}
		}
	}

	for guild in &new_guilds {
		for i in 0..guilds.len() {
			if guilds[i].id == guild.id {
				guilds.remove(i);
				break;
			}
		}
	}

	guilds.append(&mut new_guilds);

	return guilds;
}
