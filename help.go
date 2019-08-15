/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2019 Mnpn

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"fmt"
	"strings"
)

func printHelp(search string) {
	search = strings.ToLower(search)
	help := make([]string, 0)

	help = append(help, "help [search]\tShow help menu. Optionally search.")
	help = append(help, "exit\tExit DiscordConsole")
	help = append(help, "exec\tExecute a shell command")
	help = append(help, "run\tRun a LUA file with DiscordConsole's special functions")
	help = append(help, "alias <command> <new command>\tAdd a new alias for a command.")
	help = append(help, "lang <language>\tSame as starting with --lang")
	help = append(help, "permcalc [preset]\tOpen the permission calculator, and optionally with pre-set values.")
	help = append(help, "")
	if userType != typeWebhook {
		help = append(help, "guilds\t\tList guilds/servers this bot is added to.")
		help = append(help, "guild <id>\tSelect a guild to use for further commands.")
		help = append(help, "channels\tList channels in your selected guild.")
		help = append(help, "channel <id>\tSelect a channel to use for further commands.")
		help = append(help, "pchannels\tList private channels a.k.a. 'DMs'.")
		help = append(help, "vchannels\tList voice channels in your selected guild.")
		help = append(help, "dm <user id>\tCreate a DM with specific user.")
		help = append(help, "")
		help = append(help, "region <list/set> (<region>)\tSet current guild region.")
		help = append(help, "")
		help = append(help, "info <user/guild/channel/settings> (for user: <id/@me>) [property] (or info u/g/c/s)\tGet information about a user, server, channel or your set Discord settings!")
		help = append(help, "read <message id> [property]\tRead or get info from a message. Properties: (empty), text, channel, timestamp, author, "+
			"author_email, author_name, author_avatar, author_bot, embed; 'cache' may be used as message ID.")
		help = append(help, "pin <message id>\tPin a message to the current channel.")
		help = append(help, "unpin <message id>\tUnpin a message from the current channel.")
		help = append(help, "")
	}
	help = append(help, "say <stuff>\tSend a message in your selected channel. `say toggle` starts chat-mode, and `toggle` ends it.")
	if userType != typeWebhook {
		help = append(help, "${Placeholders}:\tReplaces e.g. ${u.name} with "+userObj.Username+".")
	}
	help = append(help, "sayfile <path>\tSend the contents of a file (auto-splitted).")
	help = append(help, "big <stuff>\tSend a message, but attempt to make it using emojis!")
	help = append(help, "embed <json>\tSend an embed! (ADVANCED!) See https://discordapp.com/developers/docs/resources/channel#embed-object")
	if userType != typeWebhook {
		help = append(help, "tts <stuff>\tSend a TTS message in your selected channel.")
		help = append(help, "file <file>\tUpload file to selected channel.")
		help = append(help, "")
		help = append(help, "edit <message id> <stuff>\tEdit a message in your selected channel.")
		help = append(help, "editembed <message id> <json>\tEdit a message embed in your selected channel.")
		help = append(help, "del <message id>\tDelete a message in the selected channel.")
		help = append(help, "delall [since message id]\tBulk delete messages since a specific message")
		help = append(help, "log <directly/file> <amount OR filename>\tLog the last few messages in console or to a file.")
		help = append(help, "react add/del <message id> <emoji unicode/id>\tReact to a message")
		help = append(help, "react big <message id> <stuff>\tLike the 'big' command, but in reactions!")
		help = append(help, "react delall <message id>\tDelete all reactions")
		help = append(help, "")
		help = append(help, "playing [game]\tSet your playing status. Run without an argument to clear.")
		help = append(help, "streaming [url] [game]\tSet your streaming status.")
		help = append(help, "game <streaming/watching/listening> <name> [details] [extra text]\tSet a custom status.")
		help = append(help, "typing\tSimulate typing in selected channel.")
		help = append(help, "")
		help = append(help, "members\tList (max 100) members in selected guild")
		help = append(help, "invite create [expires] [max uses] ['temp'] OR invite accept <code> OR invite read <code> OR invite list OR invite revoke <code>\tCreate an invite, accept an existing one, see invite information, list all invites or revoke an invite.")
		help = append(help, "")
		help = append(help, "role list\tList all roles in selected guild.")
		help = append(help, "role add <user id> <role id>\tAdd role to user")
		help = append(help, "role rem <user id> <role id>\tRemove role from user")
		help = append(help, "role create\tCreate new role")
		help = append(help, "role edit <role id> <flag> <value>\tEdit a role. Flags are: name, color, separate, perms, mention")
		help = append(help, "role delete <role id>\tDelete a role.")
		help = append(help, "")
		help = append(help, "nick <id> [nickname]\tChange somebody's nickname")
		help = append(help, "nickall [nickname]\tChange ALL nicknames!")
		help = append(help, "")
		help = append(help, "messages [scope]\tIntercepting messages. Optionally, scope can have a filter on it: all, mentions, private, "+
			"current (default), none")
		help = append(help, "intercept [yes/no]\tToggle intercepting 'console.' commands in Discord.")
		help = append(help, "output [yes/no]\tToggle showing 'console.' outputs directly in Discord.")
		help = append(help, "back\tJump to previous guild and/or channel.")
		help = append(help, "")
		help = append(help, "new <channel/vchannel/guild/category> <name>\tCreate a new guild or channel")
		help = append(help, "bans\tList all bans")
		help = append(help, "ban <user id> <optional reason>\tBan user")
		help = append(help, "unban <user id>\tUnban user")
		help = append(help, "kick <user id> <optional reason>\tKick user")
		help = append(help, "leave\tLeave selected guild!")
		help = append(help, "ownership <id>\tTransfer ownership.")
		help = append(help, "")
		// Deleting a category also works when selecting channel, but this is less confusing, I hope.
		help = append(help, "delete <guild/channel/category> <id>\tDelete a channel, guild or category.")
		help = append(help, "")
		help = append(help, "play <dca audio file>\tPlays a song in the selected voice channel")
		help = append(help, "stop\tStops playing any song.")
		help = append(help, "move <user id> <vchannel id>\tMove a user to another voice channel.")
		help = append(help, "")
	}
	help = append(help, "name <name>\tChange username completely.")
	help = append(help, "avatar <link/file>\tChange avatar to a link or file.")
	if userType != typeWebhook {
		help = append(help, "status <value>\tSet the user status. Possible values are: online, idle, dnd and invisible.")
		help = append(help, "")
		help = append(help, "friend <add/accept/remove/list> (<name>)\tManage your friends. Add, accept, remove and list them.")
		help = append(help, "block <user id>\tBlock a user.")
		help = append(help, "")
		help = append(help, "bookmarks\tList all bookmarks in the console.")
		help = append(help, "bookmark <name>\tCreate new bookmark out of current location. If the name starts with -, it removes the bookmark.")
		help = append(help, "go <bookmark>\tJump to the specified bookmark.")
		help = append(help, "")
		help = append(help, "rl [full]\tReload cache. If 'full' is set, it also restarts the session.")
	}

	if search != "" {
		help2 := make([]string, 0)
		for _, line := range help {
			if strings.Contains(strings.ToLower(line), search) {
				help2 = append(help2, line)
			}
		}
		help = help2
	}

	fmt.Println(strings.Join(help, "\n"))
}
