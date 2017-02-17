package main;

import (
	"fmt"
)

func PrintHelp(){
	fmt.Println("help\tShow help menu");
	fmt.Println("exit\tExit DiscordConsole");
	fmt.Println("exec\tExecute a shell command");
	fmt.Println();
	fmt.Println("guilds\tList guilds/servers this bot is added to.");
	fmt.Println("guild <id>\tSelect a guild to use for further commands.");
	fmt.Println("channels\tList channels in your selected guild.");
	fmt.Println("channel <id>\tSelect a channel to use for further commands.");
	fmt.Println("general\tSelect the 'general' channel for further commands.");
	fmt.Println("pchannels\tList private channels a.k.a. 'DMs'.");
	fmt.Println("cinfo <property>\tGet information about channel. Properties: guild, name, topic, type");
	fmt.Println("dm <user id>\tCreate a DM with specific user.");
	fmt.Println();
	fmt.Println("say <stuff>\tSend a message in your selected channel.");
	fmt.Println("embed <json>\tSend an embed! (ADVANCED!) See https://discordapp.com/developers/docs/resources/channel#embed-object");
	fmt.Println("file <file>\tUpload file to selected channel.");
	fmt.Println("edit <message id> <stuff>\tEdit a message in your selected channel.");
	fmt.Println("del <message id>\tDelete a message in the selected channel.");
	fmt.Println("delall [since message id]\tBulk delete messages since a specific message");
	fmt.Println("read <message id> [property]\tRead or get info from a message. Properties: (empty), text, author, channel");
	fmt.Println("log [output file]\tLog the last few messages in console or to a file.");
	fmt.Println();
	fmt.Println("playing [game]\tSet your playing status.");
	fmt.Println("streaming [url] [game]\tSet your streaming status");
	fmt.Println("typing\tSimulate typing in selected channel...");
	fmt.Println();
	fmt.Println("members\tList (max 100) members in selected guild");
	fmt.Println("invite [code]\tCreate (permanent) instant invite, or accept an incoming one.");
	fmt.Println();
	fmt.Println("roles\tList all roles in selected guild.");
	fmt.Println("roleadd <user id> <role id>\tAdd role to user");
	fmt.Println("roledel <user id> <role id>\tRemove role from user");
	fmt.Println("rolecreate\tCreate new role");
	fmt.Println("roleedit <role id> <flag> <value>\tEdit a role. Flags are: name, color, separate, perms, mention");
	fmt.Println("roledelete <role id>\tDelete a role.");
	fmt.Println();
	fmt.Println("nick [nickname]\tChange own nicknakme");
	fmt.Println("nickall [nickname]\tChange ALL nicknames!");
	fmt.Println();
	fmt.Println("enablemessages\tEnable intercepting messages");
	fmt.Println("disablemessages\tReverts the above.");
	fmt.Println("reply\tJump to the channel of the last received message.");
	fmt.Println("back\tJump to previous guild and/or channel.");
	fmt.Println();
	fmt.Println("bans\tList all bans");
	fmt.Println("ban <user id>\tBan user");
	fmt.Println("unban <user id>\tUnban user");
	fmt.Println("kick <user id>\tKick user");
	fmt.Println("leave\tLeave selected guild!");
}
