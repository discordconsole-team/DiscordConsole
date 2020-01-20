/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2020 Mnpn

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
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/atotto/clipboard"
	"github.com/bwmarrin/discordgo"
	"github.com/discordconsole-team/DiscordConsole/PermCalc"
	"github.com/fatih/color"
	"github.com/jD91mZM2/gtable"
	"github.com/jD91mZM2/stdutil"
)

var mutexCommand sync.Mutex

var lastUsedMsg string
var lastUsedRole string

var cacheRead *discordgo.Message
var cacheUser []*keyval
var cacheInvite []*keyval

var messages = messagesNone
var intercept = true
var output = false

var aliases map[string]string
var clear map[string]func()

var webhookCommands = []string{"help", "big", "say", "sayfile", "embed", "name", "avatar", "exit", "exec", "run", "lang"}

func command(session *discordgo.Session, source commandSource, cmd string, w io.Writer) (returnVal string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}

	parts, err := parse(substitute, cmd)

	if err != nil {
		stdutil.PrintErr(err.Error(), nil)
		return
	}

	cmd = parts[0]
	args := parts[1:]

	returnVal = commandRaw(session, source, cmd, args, w)
	return
}

func commandRaw(session *discordgo.Session, source commandSource, cmd string, args []string, w io.Writer) (returnVal string) {
	defer handleCrash()
	nargs := len(args)

	if !source.NoMutex {
		mutexCommand.Lock()
		defer mutexCommand.Unlock()
	}

	if aliascmd, ok := aliases[cmd]; ok && !source.Alias && cmd != "alias" {
		if nargs >= 1 {
			aliascmd += " " + strings.Join(args, " ")
		}
		colors := w == color.Output
		if colors {
			colorAutomated.Set()
		}
		writeln(w, aliascmd)
		if colors {
			color.Unset()
		}

		// Won't use source anywhere else.
		// No reason to copy the variable.
		source.Alias = true
		source.NoMutex = true
		return command(session, source, aliascmd, w)
	}

	if userType == typeWebhook {
		allowed := false
		for _, allow := range webhookCommands {
			if cmd == allow {
				allowed = true
			}
		}

		if !allowed {
			stdutil.PrintErr(tl("invalid.webhook.command"), nil)
			return
		}
	}

	switch cmd {
	case "help":
		search := strings.Join(args, " ")
		printHelp(search)
	case "exit":
		closing = true
	case "exec":
		if nargs < 1 {
			stdutil.PrintErr("exec <command>", nil)
			return
		}

		cmd := strings.Join(args, " ")

		err := execute(sh, c, cmd)
		if err != nil {
			stdutil.PrintErr(tl("failed.exec"), err)
		}
	case "run":
		if nargs < 1 {
			stdutil.PrintErr("run <lua script>", nil)
			return
		}
		var script string
		var scriptArgs []string

		scriptName := true
		for i, arg := range args {
			if scriptName {
				if i != 0 {
					script += " "
				}
				if strings.HasSuffix(arg, ":") {
					scriptName = false
					arg = arg[:len(arg)-1]
				}
				script += arg
			} else {
				scriptArgs = append(scriptArgs, arg)
			}
		}

		err := fixPath(&script)
		if err != nil {
			stdutil.PrintErr(tl("failed.fixpath"), err)
		}

		mutexCommand.Unlock()
		err = runLua(session, script, scriptArgs...)
		mutexCommand.Lock()
		if err != nil {
			stdutil.PrintErr(tl("failed.lua.run"), err)
		}
	case "lang":
		if nargs < 1 {
			stdutil.PrintErr("lang <language>", nil)
			return
		}

		loadLangAuto(args[0])
	case "guilds", "guild", "servers", "server", "channels", "pchannels", "vchannels",
		"channel", "dm", "bookmarks", "bookmark", "go":
		returnVal = commandsNavigate(session, cmd, args, nargs, w)
	case "say", "tts", "embed", "quote", "big", "file", "edit",
		"editembed", "sayfile", "del", "delall":
		returnVal = commandsSay(session, source, cmd, args, nargs, w)
	case "log":
		if nargs < 2 {
			stdutil.PrintErr("log <directly/file> <number OR filename>", nil)
			return
		}

		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		var file io.Writer
		directly := false
		limit := 100 // 100 is Discord's limit.

		switch strings.ToLower(args[0]) {
		case "directly":
			directly = true
			file = w
			var err error
			limit, err = strconv.Atoi(args[1])
			if err != nil {
				stdutil.PrintErr(tl("invalid.number"), nil)
				return
			}
		case "file":
			name := strings.Join(args[1:], " ")
			err := fixPath(&name)
			if err != nil {
				stdutil.PrintErr(tl("failed.fixpath"), err)
			}

			file2, err := os.Create(name)
			if err != nil {
				stdutil.PrintErr(tl("failed.file.open"), err)
				return
			}
			defer file2.Close()

			file = file2
		default:
			stdutil.PrintErr("log <directly/file> <number OR filename>", nil)
			return
		}

		msgs, err := session.ChannelMessages(loc.channel.ID, limit, "", "", "")
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.query"), err)
			return
		}

		for i := len(msgs) - 1; i >= 0; i-- {
			msg := msgs[i]
			if msg.Author == nil {
				return
			}
			s := ""
			if directly {
				s = "(ID " + msg.ID + ") "
			}

			err = writeln(file, s+msg.Author.Username+"#"+msg.Author.Discriminator+": "+msgToString(msg))
			if err != nil && !directly {
				stdutil.PrintErr(tl("failed.msg.write"), err)
				return
			}
		}
	case "members":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		members, err := session.GuildMembers(loc.guild.ID, "", 100)
		if err != nil {
			stdutil.PrintErr(tl("failed.members"), err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name", "Nick")

		for _, member := range members {
			table.AddRow()
			table.AddStrings(member.User.ID, member.User.String(), member.Nick)
		}
		writeln(w, table.String())
	case "invite":
		if nargs < 1 {
			stdutil.PrintErr("invite\taccept <code>\n\tcreate [expire] [max uses] ['temp']\n\tlist\n\trevoke <code>", nil)
			return
		}
		switch strings.ToLower(args[0]) {
		case "see", "read":
			if nargs < 2 {
				stdutil.PrintErr("invite see <code> [property]", nil)
				return
			}

			var keyvals []*keyval
			if strings.EqualFold(args[1], "cache") {
				if cacheInvite == nil {
					stdutil.PrintErr(tl("invalid.cache"), nil)
					return
				}

				keyvals = cacheInvite
			} else {
				invite, err := session.Invite(args[1])
				if err != nil {
					stdutil.PrintErr(tl("failed.invite"), err)
					return
				}

				keyvals = invite2array(invite)
				cacheInvite = keyvals
			}

			if nargs < 3 {
				for _, keyval := range keyvals {
					writeln(w, keyval.String())
				}
			} else {
				var ok bool
				returnVal, ok = findValByKey(keyvals, args[2])
				if !ok {
					stdutil.PrintErr(tl("invalid.value"), nil)
					return
				}

				writeln(w, returnVal)
			}
		case "accept":
			if nargs < 2 {
				stdutil.PrintErr("invite accept <code>", nil)
				return
			}
			if userType != typeUser {
				stdutil.PrintErr(tl("invalid.onlyfor.users"), nil)
				return
			}

			invite, err := session.InviteAccept(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.invite.accept"), err)
				return
			}
			writeln(w, tl("status.invite.accept"))

			loc.push(invite.Guild, invite.Channel)
		case "create":
			if loc.channel == nil {
				stdutil.PrintErr(tl("failed.channel"), nil)
				return
			}

			// Restricting this to bots. Discord resends a verification email
			// and it errors with "json unmarshal" when attempted from a user account.
			// I'd rather not have that.
			if userType != typeBot {
				stdutil.PrintErr(tl("invalid.onlyfor.bots"), nil)
				return
			}

			inviteObj := discordgo.Invite{}
			if nargs >= 2 {
				min, err := strconv.Atoi(args[1])
				if err != nil {
					stdutil.PrintErr(tl("invalid.number"), nil)
					return
				}
				inviteObj.MaxAge = 60 * min
				if nargs >= 3 {
					num, err := strconv.Atoi(args[2])
					if err != nil {
						stdutil.PrintErr(tl("invalid.number"), nil)
						return
					}
					inviteObj.MaxUses = num

					if nargs >= 4 && strings.EqualFold(args[3], "temp") {
						inviteObj.Temporary = true
					}
				}
			}

			invite, err := session.ChannelInviteCreate(loc.channel.ID, inviteObj)
			if err != nil {
				stdutil.PrintErr(tl("failed.invite.create"), err)
				return
			}
			writeln(w, tl("status.invite.create")+" "+invite.Code)
			returnVal = invite.Code
		case "list":
			if loc.guild == nil {
				stdutil.PrintErr(tl("invalid.guild"), nil)
				return
			}

			invites, err := session.GuildInvites(loc.guild.ID)
			if err != nil {
				stdutil.PrintErr(tl("failed.invite"), err)
				return
			}

			table := gtable.NewStringTable()
			table.AddStrings("Invites")

			for _, invite := range invites {
				table.AddRow()
				table.AddStrings(invite.Code)
			}

			writeln(w, table.String())
		case "revoke":
			if nargs < 2 {
				stdutil.PrintErr("invite revoke <code>", nil)
				return
			}

			invite, err := session.InviteDelete(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.revoke"), err)
				return
			}
			writeln(w, tl("information.revoked.successfully")+invite.Code)
		default:
			stdutil.PrintErr(tl("invalid.value"), nil)
		}
	case "messages":
		if len(args) < 1 {
			messages = messagesCurrent
			return
		}

		val, ok := typeMessages[strings.ToLower(args[0])]
		if !ok {
			stdutil.PrintErr(tl("invalid.value"), nil)
			return
		}
		messages = val
	case "intercept":
		if len(args) < 1 {
			intercept = !intercept
			returnVal = strconv.FormatBool(intercept)
			writeln(w, returnVal)
			return
		}

		state, err := parseBool(args[0])
		if err != nil {
			stdutil.PrintErr("", err)
			return
		}
		intercept = state
	case "output":
		if len(args) < 1 {
			output = !output
			returnVal = strconv.FormatBool(output)
			writeln(w, returnVal)
			return
		}

		state, err := parseBool(args[0])
		if err != nil {
			stdutil.PrintErr("", err)
			return
		}
		output = state
	case "back":
		loc.push(lastLoc.guild, lastLoc.channel)
	case "ban":
		if nargs < 1 {
			stdutil.PrintErr("ban <user id> [description]", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		var err error
		if nargs < 2 {
			err = session.GuildBanCreate(loc.guild.ID, args[0], 0)
		} else {
			err = session.GuildBanCreateWithReason(loc.guild.ID, args[0], strings.Join(args[1:], " "), 0)
		}
		if err != nil {
			stdutil.PrintErr(tl("failed.ban.create"), err)
		}
	case "unban":
		if nargs < 1 {
			stdutil.PrintErr("unban <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildBanDelete(loc.guild.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.ban.delete"), err)
		}
	case "kick":
		if nargs < 1 {
			stdutil.PrintErr("kick <user id> [reason]", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		var err error
		if nargs < 2 {
			err = session.GuildMemberDelete(loc.guild.ID, args[0])
		} else {
			err = session.GuildMemberDeleteWithReason(loc.guild.ID, args[0], strings.Join(args[1:], " "))
		}
		if err != nil {
			stdutil.PrintErr(tl("failed.kick"), err)
		}
	case "leave":
		var err error
		if nargs < 1 {
			if loc.guild == nil {
				stdutil.PrintErr(tl("invalid.guild"), nil)
				return
			}
			err = session.GuildLeave(loc.guild.ID)
		} else {
			err = session.GuildLeave(args[0])
		}

		if err != nil {
			stdutil.PrintErr(tl("failed.leave"), err)
			return
		}

		loc.push(nil, nil)
	case "bans":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}
		arg := ""
		if nargs > 0 {
			arg = strings.ToLower(args[0])
			if arg != "text" {
				stdutil.PrintErr("bans [text]", nil)
				return
			}
		}

		bans, err := session.GuildBans(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.ban.list"), err)
			return
		}

		if nargs < 1 {
			lengthNotify := false
			table := gtable.NewStringTable()
			table.AddStrings("User ID", "Username", "Reason")

			for _, ban := range bans {
				reason := ban.Reason
				if len(reason) > 100 {
					reason = reason[0:100] + "..."
					lengthNotify = true
				}
				table.AddRow()
				table.AddStrings(ban.User.ID, ban.User.Username, reason)
			}

			writeln(w, table.String())
			if lengthNotify {
				fmt.Println(tl("information.length") + " \"bans text\".")
			}
		} else if arg == "text" { // only reason that this exists is because gtable can break if the reason is too long
			for _, ban := range bans {
				reason := ""
				if ban.Reason != "" {
					reason = "): " + ban.Reason
				} else {
					reason = ")"
				}
				writeln(w, ban.User.Username + "#" + ban.User.Discriminator + " (" + ban.User.ID + reason + "\n")
			}
		}
	case "nickall":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		members, err := session.GuildMembers(loc.guild.ID, "", 100)
		if err != nil {
			stdutil.PrintErr(tl("failed.members"), err)
			return
		}

		nick := strings.Join(args, " ")

		for _, member := range members {
			err := session.GuildMemberNickname(loc.guild.ID, member.User.ID, nick)
			if err != nil {
				stdutil.PrintErr(tl("failed.nick"), err)
			}
		}
	case "play":
		if userType != typeBot {
			stdutil.PrintErr(tl("invalid.onlyfor.bots"), nil)
			return
		}
		if nargs < 1 {
			stdutil.PrintErr("play <dca audio file>", nil)
			return
		}
		if vc == nil {
			stdutil.PrintErr(tl("invalid.channel.voice"), nil)
			return
		}
		if playing != "" {
			stdutil.PrintErr(tl("invalid.music.playing"), nil)
			return
		}

		file := strings.Join(args, " ")
		err := fixPath(&file)
		if err != nil {
			stdutil.PrintErr(tl("failed.fixpath"), err)
		}

		playing = file

		writeln(w, tl("status.loading"))

		var buffer [][]byte
		err = loadAudio(file, &buffer)
		if err != nil {
			stdutil.PrintErr(tl("failed.file.load"), err)
			playing = ""
			return
		}

		writeln(w, "Loaded!")
		writeln(w, "Playing!")

		go func(buffer [][]byte, session *discordgo.Session, guild, channel string) {
			play(buffer, session, guild, channel)
			playing = ""
		}(buffer, session, loc.guild.ID, loc.channel.ID)
	case "stop":
		if userType != typeBot {
			stdutil.PrintErr(tl("invalid.onlyfor.bots"), nil)
			return
		}
		playing = ""
	case "react":
		if nargs < 2 {
			stdutil.PrintErr("react\tadd/del <message id> <emoji unicode/id>\n\tdelall <message id>\n\tbig <message id> <text>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		switch strings.ToLower(args[0]) {
		case "add":
			fallthrough
		case "del":
			if nargs < 3 {
				stdutil.PrintErr("react add/del <message id> <emoji unicode/id>", nil)
				return
			}
			if args[0] == "add" {
				err := session.MessageReactionAdd(loc.channel.ID, args[1], args[2])
				if err != nil {
					stdutil.PrintErr(tl("failed.react"), err)
					return
				}
			} else {
				err := session.MessageReactionRemove(loc.channel.ID, args[1], args[2], "@me")
				if err != nil {
					stdutil.PrintErr(tl("failed.react.del"), err)
					return
				}
			}
		case "delall":
			err := session.MessageReactionsRemoveAll(loc.channel.ID, args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.react.delall"), err)
				return
			}
		case "big":
			used := ""

			for _, c := range strings.Join(args[2:], " ") {
				str := string(toEmoji(c))

				if strings.Contains(used, str) {
					writeln(w, tl("failed.react.used"))
					continue
				}
				used += str

				err := session.MessageReactionAdd(loc.channel.ID, args[1], str)
				if err != nil {
					stdutil.PrintErr(tl("failed.react"), err)
				}
			}
		default:
			stdutil.PrintErr("react\tadd/del <message id> <emoji unicode/id>\n\tdelall <message id>\n\tbig <message id> <text>", nil)
		}
	case "block":
		if nargs < 1 {
			stdutil.PrintErr("block <user id>", nil)
			return
		}
		if userType != typeUser {
			stdutil.PrintErr(tl("invalid.onlyfor.users"), nil)
			return
		}
		err := session.RelationshipUserBlock(args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.block"), err)
			return
		}
	case "friend":
		if userType != typeUser {
			stdutil.PrintErr(tl("invalid.onlyfor.users"), nil)
			return
		}
		if nargs < 1 {
			stdutil.PrintErr("friend <add/accept/remove/list> <user id>", nil)
			return
		}
		switch strings.ToLower(args[0]) {
		case "list":
			relations, err := session.RelationshipsGet()
			if err != nil {
				stdutil.PrintErr(tl("failed.friend.list"), err)
				return
			}

			table := gtable.NewStringTable()
			table.AddStrings("ID", "Type", "Name")

			for _, relation := range relations {
				table.AddRow()
				table.AddStrings(relation.ID, typeRelationships[relation.Type], relation.User.Username)
			}

			writeln(w, table.String())
		case "add":
			if nargs < 2 {
				stdutil.PrintErr("friend add <user id>", nil)
				return
			}
			err := session.RelationshipFriendRequestSend(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.friend.add"), err)
				return
			}
		case "accept":
			if nargs < 2 {
				stdutil.PrintErr("friend accept <user id>", nil)
				return
			}
			err := session.RelationshipFriendRequestAccept(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.friend.add"), err)
				return
			}
		case "remove":
			if nargs < 2 {
				stdutil.PrintErr("friend remove <user id>", nil)
				return
			}
			err := session.RelationshipDelete(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.friend.remove"), err)
				return
			}
		default:
			stdutil.PrintErr("friend <add/accept/remove/list> <name>", nil)
		}
	case "rl":
		full := nargs >= 1 && strings.EqualFold(args[0], "full")

		var err error
		if full {
			writeln(w, tl("rl.session"))
			session.Close()
			err = session.Open()
			if err != nil {
				stdutil.PrintErr(tl("failed.session.start"), err)
			}
		}

		writeln(w, tl("rl.cache.loc"))
		var guild *discordgo.Guild
		var channel *discordgo.Channel

		if loc.guild != nil {
			guild, err = session.Guild(loc.guild.ID)

			if err != nil {
				stdutil.PrintErr(tl("failed.guild"), err)
				return
			}
		}

		if loc.channel != nil {
			channel, err = session.Channel(loc.channel.ID)

			if err != nil {
				stdutil.PrintErr(tl("failed.channel"), err)
				return
			}
		}

		loc.guild = guild
		loc.channel = channel
		pointerCache = ""

		writeln(w, tl("rl.cache.vars"))
		cacheGuilds = nil
		cacheChannels = nil
		cacheAudio = make(map[string][][]byte)
		bookmarksCache = make(map[string]*location)

		lastLoc = &location{}
		lastUsedMsg = ""
		lastUsedRole = ""

		cacheRead = nil
		cacheUser = nil
	case "avatar", "name", "playing", "streaming", "typing", "nick", "status", "game":
		returnVal = commandsUserMod(session, cmd, args, nargs, w)
	case "read", "info":
		returnVal = commandsQuery(session, cmd, args, nargs, w)
	case "role":
		returnVal = commandsRoles(session, cmd, args, nargs, w)
	case "region":
		if nargs < 1 {
			stdutil.PrintErr("region list OR region set <region>", nil)
			return
		}
		switch strings.ToLower(args[0]) {
		case "list":
			regions, err := session.VoiceRegions()
			if err != nil {
				stdutil.PrintErr(tl("failed.voice.regions"), err)
				return
			}

			table := gtable.NewStringTable()
			table.AddStrings("ID", "Name", "Port")

			for _, region := range regions {
				table.AddRow()
				table.AddStrings(region.ID, region.Name, strconv.Itoa(region.Port))
			}

			writeln(w, table.String())
		case "set":
			if nargs < 2 {
				stdutil.PrintErr("region set <region>", nil)
				return
			}
			if loc.guild == nil {
				stdutil.PrintErr(tl("invalid.guild"), nil)
				return
			}

			_, err := session.GuildEdit(loc.guild.ID, discordgo.GuildParams{
				Region: args[1],
			})
			if err != nil {
				stdutil.PrintErr(tl("failed.guild.edit"), err)
			}
		}
	case "alias":
		if nargs <= 0 {
			for alias, aliascmd := range aliases {
				writeln(w, alias+"=`"+aliascmd+"`")
			}
		} else if nargs == 1 {
			delete(aliases, strings.ToLower(args[0]))
		} else {
			if aliases == nil {
				aliases = make(map[string]string)
			}

			aliases[strings.ToLower(args[0])] = strings.Join(args[1:], " ")
		}
	case "ownership":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		if nargs < 1 {
			stdutil.PrintErr("ownership <user id>", nil)
			return
		}
		id := args[0]

		if loc.guild.OwnerID != userObj.ID {
			stdutil.PrintErr(tl("invalid.not.owner"), nil)
			return
		}

		member, err := session.State.Member(loc.guild.ID, id)
		if err != nil {
			stdutil.PrintErr(tl("failed.user"), err)
			return
		}

		if userType == typeBot {
			stdutil.PrintErr(tl("invalid.onlyfor.users"), nil)
			return
		}

		execerr := execute("clear")
		if execerr != nil {
			stdutil.PrintErr("", err)
		}
		// We're Microsoft, and we're special!
		// We need our own damn part, because fuck you!
		// Windows is being a donkey. We'll only be clearing on Unix.

		c := color.New(color.FgHiRed)
		c.Println(tl("information.wait"))
		fmt.Println(tl("information.give.ownership") + member.User.Username + "#" + member.User.Discriminator + ". " + tl("information.irreversible"))
		fmt.Println(tl("information.confirmation") + " (y/n)")

		var response string
		_, err1 := fmt.Scanln(&response)
		if err1 != nil {
			stdutil.PrintErr("", err1)
		}

		state, err := parseBool(response)
		if err != nil {
			stdutil.PrintErr("", err)
			return
		}

		if state == true {
			_, oerr := session.GuildEdit(loc.guild.ID, discordgo.GuildParams{OwnerID: id})
			if oerr != nil {
				stdutil.PrintErr(tl("failed.transfer"), oerr)
				return
			}
		} else {
			fmt.Println(tl("information.aborted"))
			return
		}
	case "permcalc":
		if !source.Terminal {
			stdutil.PrintErr(tl("invalid.source.terminal"), nil)
			return
		}
		pm := permcalc.PermCalc{}

		if nargs >= 1 {
			i, err := strconv.Atoi(args[0])
			if err != nil {
				stdutil.PrintErr(tl("invalid.number"), nil)
				return
			}
			pm.Perm = i
		}

		err := pm.Show()
		if err != nil {
			stdutil.PrintErr("failed.permcalc", err)
			return
		}
		writeln(w, strconv.Itoa(pm.Perm))
	case "crash":
		if nargs >= 1 && strings.EqualFold(args[0], "die") {
			// Make error async for more "damage"
			go func() {
				panic("die")
			}()
			return
		}
		panic("triggered crash")
	case "pin":
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		if nargs < 1 {
			stdutil.PrintErr("pin <message id>", nil)
			return
		}

		err := session.ChannelMessagePin(loc.channel.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.pin"), err)
			return
		}
	case "unpin":
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		if nargs < 1 {
			stdutil.PrintErr("unpin <message id>", nil)
			return
		}

		err := session.ChannelMessageUnpin(loc.channel.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.unpin"), err)
			return
		}
	case "new":
		if nargs < 2 {
			stdutil.PrintErr("new <channel/vchannel/guild/category> <name>", nil)
			return
		}
		switch strings.ToLower(args[0]) {
		case "guild":
			g, err := session.GuildCreate(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.guild.create"), err)
				return
			}
			fmt.Println(tl("information.guild") + args[1] + tl("information.created.successfully") + g.ID + ".")
		case "channel":
			if loc.guild == nil {
				stdutil.PrintErr(tl("invalid.guild"), nil)
				return
			}
			c, err := session.GuildChannelCreate(loc.guild.ID, args[1], discordgo.ChannelTypeGuildText)
			if err != nil {
				stdutil.PrintErr(tl("failed.channel.create"), err)
				return
			}
			fmt.Println(tl("information.channel") + args[1] + tl("information.created.successfully") + c.ID + ".")
		case "vchannel":
			if loc.guild == nil {
				stdutil.PrintErr(tl("invalid.guild"), nil)
				return
			}
			vc, err := session.GuildChannelCreate(loc.guild.ID, args[1], discordgo.ChannelTypeGuildVoice)
			if err != nil {
				stdutil.PrintErr(tl("failed.channel.create"), err)
				return
			}
			fmt.Println(tl("information.channel") + args[1] + tl("information.created.successfully") + vc.ID + ".")
		case "category":
			if loc.guild == nil {
				stdutil.PrintErr(tl("invalid.guild"), nil)
				return
			}
			c, err := session.GuildChannelCreate(loc.guild.ID, args[1], discordgo.ChannelTypeGuildCategory)
			if err != nil {
				stdutil.PrintErr(tl("failed.category.create"), err)
				return
			}
			fmt.Println(tl("information.category") + args[1] + tl("information.created.successfully") + c.ID + ".")
		default:
			stdutil.PrintErr("new <channel/vchannel/guild/category> <name>", nil)
		}
	case "delete":
		if nargs < 2 {
			stdutil.PrintErr("delete <guild/channel/category> <id>", nil)
			return
		}
		switch strings.ToLower(args[0]) {
		case "channel", "category":
			_, err := session.ChannelDelete(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.channel.delete"), err)
				return
			}
			fmt.Println(tl("information.channel") + args[1] + tl("information.deleted.successfully"))
		case "guild":
			_, err := session.GuildDelete(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.guild.delete"), err)
				return
			}
			fmt.Println(tl("information.guild") + args[1] + tl("information.deleted.successfully"))
		}
	case "move":
		if nargs < 2 {
			stdutil.PrintErr("move <user id> <vchannel id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}
		err := session.GuildMemberMove(loc.guild.ID, args[0], args[1])
		if err != nil {
			stdutil.PrintErr(tl("failed.move"), err)
			return
		}
		writeln(w, tl("information.moved"))
	case "note":
		if nargs < 2 {
			stdutil.PrintErr("note <user id> <note>", nil)
			return
		}
		err := session.UserNoteSet(args[0], strings.Join(args[1:], " "))
		if err != nil {
			stdutil.PrintErr(tl("failed.note"), err)
			return
		}
		fmt.Println(tl("information.note"))
	case "latency":
		fmt.Println(session.HeartbeatLatency())
	default:
		stdutil.PrintErr(tl("invalid.command")+" '"+cmd+"'. "+tl("invalid.command2"), nil)
	}
	return
}

func parseBool(str string) (bool, error) {
	str = strings.ToLower(str)
	if str == "yes" || str == "true" || str == "y" {
		return true, nil
	} else if str == "no" || str == "false" || str == "n" {
		return false, nil
	}
	return false, errors.New(tl("invalid.yn"))
}

func msgToString(msg *discordgo.Message) string {
	msgc := msg.Content
	for _, attachment := range msg.Attachments {
		if len(msgc) > 0 {
			msgc += " "
		}
		msgc += attachment.URL
	}
	return msgc
}

func substitute(args string) (string, bool) {
	switch args {
	case "paste":
		clipboardcontent, err := clipboard.ReadAll()
		if err != nil {
			stdutil.PrintErr((tl("failed.paste") + err.Error()), nil)
			return "", true
		}
		return clipboardcontent, true
	case "s.id":
		if loc.guild != nil {
			return loc.guild.ID, true
		}
		return "nil", true
	case "s.owner.id":
		if loc.guild != nil {
			return loc.guild.OwnerID, true
		}
		return "nil", true
	case "s.owner.mention":
		if loc.guild != nil {
			return "<@" + loc.guild.OwnerID + ">", true
		}
		return "nil", true
	case "c.id":
		if loc.channel != nil {
			return loc.channel.ID, true
		}
		return "nil", true
	// Webhooks don't have access to this info and would crash the program.
	case "u.name":
		if userType != typeWebhook {
			return userObj.Username, true
		}
	case "u.discrim":
		if userType != typeWebhook {
			return userObj.Discriminator, true
		}
	case "u.id":
		if userType != typeWebhook {
			return userObj.ID, true
		}
	case "u.mention":
		if userType != typeWebhook {
			return "<@" + userObj.ID + ">", true
		}
	default:
		return "", false
	}
	return "", false
}
