package main;

import (
	"fmt"
	"strings"
)

func printHelp(search string){
	help := make([]string, 0);
	help = append(help, "help [search]\tShow help menu. Optionally search.");
	help = append(help, "exit\tExit DiscordConsole");
	help = append(help, "exec\tExecute a shell command");
	help = append(help, "");
	help = append(help, "guilds\tList guilds/servers this bot is added to.");
	help = append(help, "guild <id>\tSelect a guild to use for further commands.");
	help = append(help, "channels\tList channels in your selected guild.");
	help = append(help, "channel <id>\tSelect a channel to use for further commands.");
	help = append(help, "pchannels\tList private channels a.k.a. 'DMs'.");
	help = append(help, "vchannels\tList voice channels in your selected guild.");
	help = append(help, "dm <user id>\tCreate a DM with specific user.");
	help = append(help, "");
	help = append(help, "cinfo <property>\tGet information about channel. Properties: guild, name, topic, type");
	help = append(help, "ginfo <property>\tGet information about guild. Properties: name, icon, region, owner, splash, members, level");
	help = append(help, "uinfo <id> <property>\tGet information about user. Properties: id, email, name, avatar, bot");
	help = append(help, "read <message id> [property]\tRead or get info from a message. Properties: (empty), text, channel, timestamp, author, " +
						"author_email, author_name, author_avatar, author_bot; 'cache' may be used as property to save, and used later as message ID.");
	help = append(help, "");
	help = append(help, "say <stuff>\tSend a message in your selected channel.");
	help = append(help, "big <stuff>\tSend a message, but attempt to make it using emojis!");
	help = append(help, "tts <stuff>\tSend a TTS message in your selected channel.");
	help = append(help, "embed <json>\tSend an embed! (ADVANCED!) See https://discordapp.com/developers/docs/resources/channel#embed-object");
	help = append(help, "file <file>\tUpload file to selected channel.");
	help = append(help, "");
	help = append(help, "edit <message id> <stuff>\tEdit a message in your selected channel.");
	help = append(help, "del <message id>\tDelete a message in the selected channel.");
	help = append(help, "delall [since message id]\tBulk delete messages since a specific message");
	help = append(help, "log [output file]\tLog the last few messages in console or to a file.");
	help = append(help, "reactadd/reactdel <message id> <emoji unicode/id>\tReact to a message");
	help = append(help, "reactbig <message id> <stuff>\tLike the 'big' command, but in reactions!");
	help = append(help, "");
	help = append(help, "playing [game]\tSet your playing status.");
	help = append(help, "streaming [url] [game]\tSet your streaming status");
	help = append(help, "typing\tSimulate typing in selected channel...");
	help = append(help, "");
	help = append(help, "members\tList (max 100) members in selected guild");
	help = append(help, "invite [code]\tCreate (permanent) instant invite, or accept an incoming one.");
	help = append(help, "");
	help = append(help, "roles\tList all roles in selected guild.");
	help = append(help, "roleadd <user id> <role id>\tAdd role to user");
	help = append(help, "roledel <user id> <role id>\tRemove role from user");
	help = append(help, "rolecreate\tCreate new role");
	help = append(help, "roleedit <role id> <flag> <value>\tEdit a role. Flags are: name, color, separate, perms, mention");
	help = append(help, "roledelete <role id>\tDelete a role.");
	help = append(help, "");
	help = append(help, "nick <id/'me'> [nickname]\tChange somebody's nickname");
	help = append(help, "nickall [nickname]\tChange ALL nicknames!");
	help = append(help, "");
	help = append(help, "enablemessages\tEnable intercepting messages");
	help = append(help, "disablemessages\tReverts the above.");
	help = append(help, "enableintercept\tEnable intercepting 'console.' commands in Discord.");
	help = append(help, "disableintercept\tReverts the above.");
	help = append(help, "reply\tJump to the channel of the last received message.");
	help = append(help, "back\tJump to previous guild and/or channel.");
	help = append(help, "");
	help = append(help, "bans\tList all bans");
	help = append(help, "ban <user id>\tBan user");
	help = append(help, "unban <user id>\tUnban user");
	help = append(help, "kick <user id>\tKick user");
	help = append(help, "leave\tLeave selected guild!");
	help = append(help, "");
	help = append(help, "play <dca audio file>\tPlays a song in the selected voice channel");
	help = append(help, "stop\tStops playing any song.");
	help = append(help, "");
	help = append(help, "avatar <link/file>\tChange avatar to a link or file.");
	help = append(help, "friends\tList all relationships the account has.");
	help = append(help, "block <user id>\tBlock a user.");
	help = append(help, "");
	help = append(help, "bookmarks\tList all bookmarks in the console.");
	help = append(help, "bookmark <name>\tCreate new bookmark out of current location. If the name starts with -, it removes the bookmark.");
	help = append(help, "go <bookmark>\tJump to the specified bookmark.");
	help = append(help, "");
	help = append(help, "rl [full]\tReload cache. If 'full' is set, it also restarts the session.");

	if(search != ""){
		help2 := make([]string, 0);
		for _, line := range help{
			if(strings.Contains(line, search)){
				help2 = append(help2, line);
			}
		}
		help = help2;
	}

	fmt.Println(strings.Join(help, "\n"));
}
