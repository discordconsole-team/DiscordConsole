package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/gtable"
	"github.com/legolord208/stdutil"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var TypeRelationships = map[int]string{
	1: "Friend",
	2: "Blocked",
	3: "Incoming request",
	4: "Sent request",
}
var TypeVerifications = map[discordgo.VerificationLevel]string{
	discordgo.VerificationLevelNone:   "None",
	discordgo.VerificationLevelLow:    "Low",
	discordgo.VerificationLevelMedium: "Medium",
	discordgo.VerificationLevelHigh:   "High",
}
var TypeMessages = map[string]int{
	"all":      MessagesAll,
	"mentions": MessagesMentions,
	"private":  MessagesPrivate,
	"current":  MessagesCurrent,
}
var TypeStatuses = map[string]discordgo.Status{
	"online":    discordgo.StatusOnline,
	"idle":      discordgo.StatusIdle,
	"dnd":       discordgo.StatusDoNotDisturb,
	"invisible": discordgo.StatusInvisible,
	"offline":   discordgo.StatusOffline,
}

type location struct {
	guild   *discordgo.Guild
	channel *discordgo.Channel
}

func (loc *location) push(guild *discordgo.Guild, channel *discordgo.Channel) {
	if loc.guild == guild && loc.channel == channel {
		return
	}
	if guild != nil && channel != nil && loc.guild != nil && loc.channel != nil &&
		loc.guild.ID == guild.ID && loc.channel.ID == channel.ID {
		return
	}
	lastLoc = *loc

	loc.guild = guild
	loc.channel = channel
	pointerCache = ""
}

var loc location
var lastLoc location
var lastMsg location

var lastUsedMsg string
var lastUsedRole string

var cacheGuilds = make(map[string]string)
var cacheChannels = make(map[string]string)
var cacheRead *discordgo.Message

const (
	MessagesNone = iota
	MessagesCurrent
	MessagesPrivate
	MessagesMentions
	MessagesAll
)

var messages = MessagesNone
var intercept = true

var webhookCommands = []string{"big", "say", "sayfile", "embed", "name", "avatar", "exit", "exec", "run"}

func command(session *discordgo.Session, cmd string) (returnVal string) {
	if cmd == "" {
		return
	}
	parts := strings.FieldsFunc(cmd, func(c rune) bool {
		return c != '\n' && unicode.IsSpace(c)
	})

	cmd = strings.ToLower(parts[0])
	args := parts[1:]
	nargs := len(args)

	if UserType == TypeWebhook {
		allowed := false
		for _, allow := range webhookCommands {
			if cmd == allow {
				allowed = true
			}
		}

		if !allowed {
			stdutil.PrintErr(lang["invalid.webhook.command"], nil)
			return
		}
	}

	switch cmd {
	case "help":
		search := strings.Join(args, " ")
		printHelp(search)
	case "exit":
		exit(session)
	case "exec":
		if nargs < 1 {
			stdutil.PrintErr("exec <command>", nil)
			return
		}

		cmd := strings.Join(args, " ")

		err := execute(SH, C, cmd)
		if err != nil {
			stdutil.PrintErr(lang["failed.exec"], err)
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
			stdutil.PrintErr(lang["failed.fixpath"], err)
		}

		err = RunLua(session, script, scriptArgs...)
		if err != nil {
			stdutil.PrintErr(lang["failed.lua.run"], err)
		}
	case "guilds":
		guilds, err := session.UserGuilds(100, "", "")
		if err != nil {
			stdutil.PrintErr(lang["failed.guild"], err)
			return
		}

		cacheGuilds = make(map[string]string)

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name")

		for _, guild := range guilds {
			table.AddRow()
			table.AddStrings(guild.ID, guild.Name)
			cacheGuilds[strings.ToLower(guild.Name)] = guild.ID
		}

		printTable(table)
	case "guild":
		if nargs < 1 {
			stdutil.PrintErr("guild <id>", nil)
			return
		}

		guildID, ok := cacheGuilds[strings.ToLower(strings.Join(args, " "))]
		if !ok {
			guildID = args[0]
		}

		guild, err := session.Guild(guildID)
		if err != nil {
			stdutil.PrintErr(lang["failed.guild"], err)
			return
		}

		channel, err := session.Channel(guildID)
		if err != nil {
			stdutil.PrintErr(lang["failed.channel"], err)
			return
		}
		loc.push(guild, channel)
	case "channels":
		channels(session, "text")
	case "channel":
		if nargs < 1 {
			stdutil.PrintErr("channel <id>", nil)
			return
		}

		channelID, ok := cacheChannels[strings.ToLower(strings.Join(args, " "))]
		if !ok {
			channelID = args[0]
		}

		channel, err := session.Channel(channelID)
		if err != nil {
			stdutil.PrintErr(lang["failed.channel"], err)
			return
		}
		if channel.IsPrivate {
			loc.push(nil, channel)
		} else {
			if loc.guild == nil || channel.GuildID != loc.guild.ID {
				guild, err := session.Guild(channel.GuildID)

				if err != nil {
					stdutil.PrintErr(lang["failed.guild"], err)
					return
				}

				loc.push(guild, channel)
			} else {
				loc.push(loc.guild, channel)
			}
		}
	case "say":
		if nargs < 1 {
			stdutil.PrintErr("say <stuff>", nil)
			return
		}
		if loc.channel == nil && UserType != TypeWebhook {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}
		msgStr := strings.Join(args, " ")

		if len(msgStr) > MsgLimit {
			stdutil.PrintErr(lang["invalid.limit.message"], nil)
			return
		}

		if UserType == TypeWebhook {
			err := session.WebhookExecute(UserId, UserToken, false, &discordgo.WebhookParams{
				Content: msgStr,
			})
			if err != nil {
				stdutil.PrintErr(lang["failed.msg.send"], err)
				return
			}
			return
		}
		msg, err := session.ChannelMessageSend(loc.channel.ID, msgStr)
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.send"], err)
			return
		}
		fmt.Println(lang["status.msg.create"] + " " + msg.ID)
		lastUsedMsg = msg.ID
		returnVal = msg.ID
	case "edit":
		if nargs < 2 {
			stdutil.PrintErr("edit <message id> <stuff>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		msg, err := session.ChannelMessageEdit(loc.channel.ID, args[0], strings.Join(args[1:], " "))
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.edit"], err)
			return
		}
		lastUsedMsg = msg.ID
	case "del":
		if nargs < 1 {
			stdutil.PrintErr("del <message id>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		err := session.ChannelMessageDelete(loc.channel.ID, args[0])
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.delete"], err)
			return
		}
	case "log":
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		directly := nargs < 1

		limit := 100
		if directly {
			limit = 10
		}

		msgs, err := session.ChannelMessages(loc.channel.ID, limit, "", "", "")
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.query"], err)
			return
		}
		s := ""

		for i := len(msgs) - 1; i >= 0; i-- {
			msg := msgs[i]
			if msg.Author == nil {
				return
			}
			if directly {
				s += "(ID " + msg.ID + ") "
			}
			s += msg.Author.Username + ": " + msg.Content + "\n"
		}

		if directly {
			fmt.Print(s)
			returnVal = s
			return
		}

		name := strings.Join(args, " ")
		err = fixPath(&name)
		if err != nil {
			stdutil.PrintErr(lang["failed.fixpath"], err)
		}

		err = ioutil.WriteFile(name, []byte(s), 0666)
		if err != nil {
			stdutil.PrintErr(lang["failed.file.write"], err)
			return
		}
		fmt.Println("Wrote chat log to '" + name + "'.")
	case "playing":
		err := session.UpdateStatus(0, strings.Join(args, " "))
		if err != nil {
			stdutil.PrintErr(lang["failed.status"], err)
		}
	case "streaming":
		var err error
		if nargs <= 0 {
			err = session.UpdateStreamingStatus(0, "", "")
		} else if nargs < 2 {
			err = session.UpdateStreamingStatus(0, strings.Join(args[1:], " "), "")
		} else {
			err = session.UpdateStreamingStatus(0, strings.Join(args[1:], " "), args[0])
		}
		if err != nil {
			stdutil.PrintErr(lang["failed.status"], err)
		}
	case "typing":
		if loc.channel == nil {
			stdutil.PrintErr(lang["failed.channel"], nil)
			return
		}
		err := session.ChannelTyping(loc.channel.ID)
		if err != nil {
			stdutil.PrintErr(lang["failed.typing"], err)
		}
	case "pchannels":
		channels, err := session.UserChannels()
		if err != nil {
			stdutil.PrintErr(lang["failed.channel"], err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Recipient")

		for _, channel := range channels {
			table.AddRow()
			recipient := ""
			if channel.Recipient != nil {
				recipient = channel.Recipient.Username
			}
			table.AddStrings(channel.ID, recipient)
		}
		printTable(table)
	case "dm":
		if nargs < 1 {
			stdutil.PrintErr("dm <user id>", nil)
			return
		}
		channel, err := session.UserChannelCreate(args[0])
		if err != nil {
			stdutil.PrintErr(lang["failed.channel.create"], err)
			return
		}
		loc.push(nil, channel)

		fmt.Println(lang["channel.select"] + " " + channel.ID)
	case "delall":
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}
		since := ""
		if nargs >= 1 {
			since = args[0]
		}
		messages, err := session.ChannelMessages(loc.channel.ID, 100, "", since, "")
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.query"], err)
			return
		}

		ids := make([]string, len(messages))
		for i, msg := range messages {
			ids[i] = msg.ID
		}

		err = session.ChannelMessagesBulkDelete(loc.channel.ID, ids)
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.query"], err)
			return
		}
		returnVal := strconv.Itoa(len(ids))
		fmt.Println("Deleted " + returnVal + " messages!")
	case "members":
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		members, err := session.GuildMembers(loc.guild.ID, "", 100)
		if err != nil {
			stdutil.PrintErr(lang["failed.members"], err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name", "Nick")

		for _, member := range members {
			table.AddRow()
			table.AddStrings(member.User.ID, member.User.Username, member.Nick)
		}
		printTable(table)
	case "invite":
		if nargs >= 1 {
			if UserType != TypeUser {
				stdutil.PrintErr(lang["invalid.onlyfor.users"], nil)
				return
			}

			invite, err := session.InviteAccept(args[0])
			if err != nil {
				stdutil.PrintErr(lang["failed.invite.accept"], err)
				return
			}
			fmt.Println(lang["status.invite.accept"])

			loc.push(invite.Guild, invite.Channel)
		} else {
			if loc.channel == nil {
				stdutil.PrintErr(lang["failed.channel"], nil)
				return
			}
			invite, err := session.ChannelInviteCreate(loc.channel.ID, discordgo.Invite{})
			if err != nil {
				stdutil.PrintErr(lang["failed.invite.create"], err)
				return
			}
			fmt.Println(lang["status.invite.create"] + " " + invite.Code)
			returnVal = invite.Code
		}
	case "file":
		if nargs < 1 {
			stdutil.PrintErr("file <file>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}
		name := strings.Join(args, " ")
		err := fixPath(&name)
		if err != nil {
			stdutil.PrintErr(lang["failed.fixpath"], err)
		}

		file, err := os.Open(name)
		if err != nil {
			stdutil.PrintErr(lang["failed.file.open"], nil)
			return
		}
		defer file.Close()

		msg, err := session.ChannelFileSend(loc.channel.ID, filepath.Base(name), file)
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.send"], err)
			return
		}
		fmt.Println(lang["status.msg.created"] + " " + msg.ID)
		return msg.ID
	case "roles":
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		roles, err := session.GuildRoles(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(lang["failed.roles"], err)
			return
		}
		sort.Slice(roles, func(i, j int) bool {
			return roles[i].Position > roles[j].Position
		})

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name", "Permissions")

		for _, role := range roles {
			table.AddRow()
			table.AddStrings(role.ID, role.Name, strconv.Itoa(role.Permissions))
		}

		printTable(table)
	case "roleadd":
		fallthrough
	case "roledel":
		if nargs < 2 {
			stdutil.PrintErr("roleadd/del <user id> <role id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		var err error
		if cmd == "roleadd" {
			err = session.GuildMemberRoleAdd(loc.guild.ID, args[0], args[1])
		} else {
			err = session.GuildMemberRoleRemove(loc.guild.ID, args[0], args[1])
		}

		if err != nil {
			stdutil.PrintErr(lang["failed.role.change"], err)
		}
	case "nick":
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}
		if nargs < 1 {
			stdutil.PrintErr("nick <id> [nickname]", nil)
			return
		}

		who := args[0]
		if strings.EqualFold(who, "@me") {
			who = "@me/nick"
			// Should hopefully only be @me in the future.
			// See https://github.com/bwmarrin/discordgo/issues/318
		}

		err := session.GuildMemberNickname(loc.guild.ID, who, strings.Join(args[1:], " "))
		if err != nil {
			stdutil.PrintErr(lang["failed.nick"], err)
		}
	case "enablemessages":
		if len(args) < 1 {
			messages = MessagesCurrent
			return
		}

		val, ok := TypeMessages[strings.ToLower(args[0])]
		if !ok {
			stdutil.PrintErr(lang["invalid.value"], nil)
			return
		}
		messages = val
		fmt.Println(lang["status.msg.intercept"])
	case "disablemessages":
		messages = MessagesNone
		fmt.Println(lang["status.msg.nointercept"])
	case "enableintercept":
		intercept = true
		fmt.Println(lang["status.cmd.intercept"])
	case "disableintercept":
		intercept = false
		fmt.Println(lang["status.cmd.nointercept"])
	case "reply":
		loc.push(lastMsg.guild, lastMsg.channel)
	case "back":
		loc, lastLoc = lastLoc, loc
		pointerCache = ""
	case "rolecreate":
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		role, err := session.GuildRoleCreate(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(lang["failed.role.create"], err)
			return
		}
		fmt.Println("Created role with ID " + role.ID)
		lastUsedRole = role.ID
		returnVal = role.ID
	case "roleedit":
		if nargs < 3 {
			stdutil.PrintErr("roleedit <roleid> <flag> <value>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		value := strings.Join(args[2:], " ")

		roles, err := session.GuildRoles(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(lang["failed.roles"], err)
			return
		}

		var role *discordgo.Role
		for _, r := range roles {
			if r.ID == args[0] {
				role = r
				break
			}
		}
		if role == nil {
			stdutil.PrintErr(lang["invalid.role"], nil)
			return
		}

		name := role.Name
		color := int64(role.Color)
		hoist := role.Hoist
		perms := role.Permissions
		mention := role.Mentionable

		switch strings.ToLower(args[1]) {
		case "name":
			name = value
		case "color":
			value = strings.TrimPrefix(value, "#")
			color, err = strconv.ParseInt(value, 16, 0)
			if err != nil {
				stdutil.PrintErr(lang["invalid.number"], nil)
				return
			}
		case "separate":
			hoist, err = parseBool(value)
			if err != nil {
				stdutil.PrintErr(err.Error(), nil)
				return
			}
		case "perms":
			perms, err = strconv.Atoi(value)
			if err != nil {
				stdutil.PrintErr(lang["invalid.number"], nil)
				return
			}
		case "mention":
			mention, err = parseBool(value)
			if err != nil {
				stdutil.PrintErr(err.Error(), nil)
				return
			}
		default:
			stdutil.PrintErr(lang["invalid.value"], nil)
			return
		}

		role, err = session.GuildRoleEdit(loc.guild.ID, args[0], name, int(color), hoist, perms, mention)
		if err != nil {
			stdutil.PrintErr(lang["failed.role.edit"], err)
			return
		}
		lastUsedRole = role.ID
		fmt.Println("Edited role " + role.ID)
	case "roledelete":
		if nargs < 1 {
			stdutil.PrintErr("roledelete <roleid>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		err := session.GuildRoleDelete(loc.guild.ID, args[0])
		if err != nil {
			fmt.Println(lang["failed.role.delete"], err)
		}
	case "ban":
		if nargs < 1 {
			stdutil.PrintErr("ban <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		err := session.GuildBanCreate(loc.guild.ID, args[0], 0)
		if err != nil {
			stdutil.PrintErr(lang["failed.ban.create"], err)
		}
	case "unban":
		if nargs < 1 {
			stdutil.PrintErr("unban <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		err := session.GuildBanDelete(loc.guild.ID, args[0])
		if err != nil {
			stdutil.PrintErr(lang["failed.ban.delete"], err)
		}
	case "kick":
		if nargs < 1 {
			stdutil.PrintErr("kick <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		err := session.GuildMemberDelete(loc.guild.ID, args[0])
		if err != nil {
			stdutil.PrintErr(lang["failed.kick"], err)
		}
	case "leave":
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		err := session.GuildLeave(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(lang["failed.leave"], err)
			return
		}

		loc.push(nil, nil)
	case "bans":
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		bans, err := session.GuildBans(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(lang["failed.ban.list"], err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("User ID", "Username", "Reason")

		for _, ban := range bans {
			table.AddRow()
			table.AddStrings(ban.User.ID, ban.User.Username, ban.Reason)
		}

		printTable(table)
	case "nickall":
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		members, err := session.GuildMembers(loc.guild.ID, "", 100)
		if err != nil {
			stdutil.PrintErr(lang["failed.members"], err)
			return
		}

		nick := strings.Join(args, " ")

		for _, member := range members {
			err := session.GuildMemberNickname(loc.guild.ID, member.User.ID, nick)
			if err != nil {
				stdutil.PrintErr(lang["failed.nick"], err)
			}
		}
	case "embed":
		if nargs < 1 {
			stdutil.PrintErr("embed <embed json>", nil)
			return
		}
		if loc.channel == nil && UserType != TypeWebhook {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		jsonstr := strings.Join(args, " ")
		var embed = &discordgo.MessageEmbed{}

		err := json.Unmarshal([]byte(jsonstr), embed)
		if err != nil {
			stdutil.PrintErr(lang["failed.json"], err)
			return
		}

		if UserType == TypeWebhook {
			err = session.WebhookExecute(UserId, UserToken, false, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				stdutil.PrintErr(lang["failed.msg.send"], err)
				return
			}
		} else {
			msg, err := session.ChannelMessageSendEmbed(loc.channel.ID, embed)
			if err != nil {
				stdutil.PrintErr(lang["failed.msg.send"], err)
				return
			}
			fmt.Println(lang["status.msg.create"] + " " + msg.ID)
			lastUsedMsg = msg.ID
			returnVal = msg.ID
		}
	case "read":
		if nargs < 1 {
			stdutil.PrintErr("read <message id> [property]", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}
		msgID := args[0]

		var msg *discordgo.Message
		var err error
		if strings.EqualFold(msgID, "cache") {
			if cacheRead == nil {
				stdutil.PrintErr(lang["invalid.cache"], nil)
				return
			}

			msg = cacheRead
		} else {
			msg, err = getMessage(session, loc.channel.ID, msgID)
			if err != nil {
				stdutil.PrintErr(lang["failed.msg.query"], err)
				return
			}
		}

		property := ""
		if len(args) >= 2 {
			property = strings.ToLower(args[1])
		}
		switch property {
		case "":
			printMessage(session, msg, false, loc.guild, loc.channel)
		case "cache":
			cacheRead = msg
			fmt.Println(lang["status.cache"])
		case "text":
			returnVal = msg.Content
		case "channel":
			returnVal = msg.ChannelID
		case "timestamp":
			t, err := timestamp(msg)
			if err != nil {
				stdutil.PrintErr(lang["failed.timestamp"], err)
				return
			}
			returnVal = t
		case "author":
			returnVal = msg.Author.ID
		case "author_email":
			returnVal = msg.Author.Email
		case "author_name":
			returnVal = msg.Author.Username
		case "author_avatar":
			returnVal = msg.Author.Avatar
		case "author_bot":
			returnVal = strconv.FormatBool(msg.Author.Bot)
		default:
			stdutil.PrintErr(lang["invalid.value"], nil)
		}

		lastUsedMsg = msg.ID
		if returnVal != "" {
			fmt.Println(returnVal)
		}
	case "cinfo":
		if nargs < 1 {
			stdutil.PrintErr("cinfo <property>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		switch strings.ToLower(args[0]) {
		case "guild":
			returnVal = loc.channel.GuildID
		case "name":
			returnVal = loc.channel.Name
		case "topic":
			returnVal = loc.channel.Topic
		case "type":
			returnVal = loc.channel.Type
		default:
			stdutil.PrintErr(lang["invalid.value"], nil)
		}

		if returnVal != "" {
			fmt.Println(returnVal)
		}
	case "vchannels":
		channels(session, "voice")
	case "play":
		if UserType != TypeBot {
			stdutil.PrintErr(lang["invalid.onlyfor.bots"], nil)
			return
		}
		if nargs < 1 {
			stdutil.PrintErr("play <dca audio file>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}
		if playing != "" {
			stdutil.PrintErr(lang["invalid.music.playing"], nil)
			return
		}

		file := strings.Join(args, " ")
		err := fixPath(&file)
		if err != nil {
			stdutil.PrintErr(lang["failed.fixpath"], err)
		}

		playing = file

		fmt.Println(lang["status.loading"])

		var buffer [][]byte
		err = loadAudio(file, &buffer)
		if err != nil {
			stdutil.PrintErr(lang["failed.file.load"], err)
			playing = ""
			return
		}

		fmt.Println("Loaded!")
		fmt.Println("Playing!")

		go func(buffer [][]byte, session *discordgo.Session, guild, channel string) {
			play(buffer, session, guild, channel)
			playing = ""
		}(buffer, session, loc.guild.ID, loc.channel.ID)
	case "stop":
		if UserType != TypeBot {
			stdutil.PrintErr(lang["invalid.onlyfor.bots"], nil)
			return
		}
		playing = ""
	case "reactadd":
		fallthrough
	case "reactdel":
		if nargs < 2 {
			stdutil.PrintErr("reactadd/reactdel <message id> <emoji unicode/id>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		var err error
		if cmd == "reactadd" {
			err = session.MessageReactionAdd(loc.channel.ID, args[0], args[1])
		} else {
			err = session.MessageReactionRemove(loc.channel.ID, args[0], args[1], "@me")
		}
		if err != nil {
			stdutil.PrintErr(lang["failed.react"], err)
			return
		}
	case "quote":
		if nargs < 1 {
			stdutil.PrintErr("quote <message id>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		msg, err := getMessage(session, loc.channel.ID, args[0])
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.query"], err)
			return
		}

		t, err := timestamp(msg)
		if err != nil {
			stdutil.PrintErr(lang["failed.timestamp"], err)
			return
		}

		msg, err = session.ChannelMessageSendEmbed(loc.channel.ID, &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    msg.Author.Username,
				IconURL: discordgo.EndpointUserAvatar(msg.Author.ID, msg.Author.Avatar),
			},
			Description: msg.Content,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Sent " + t,
			},
		})
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.send"], err)
			return
		}
		fmt.Println("Created message with ID " + msg.ID)
		lastUsedMsg = msg.ID
		returnVal = msg.ID
	case "block":
		if nargs < 1 {
			stdutil.PrintErr("block <user id>", nil)
			return
		}
		if UserType != TypeUser {
			stdutil.PrintErr(lang["invalid.onlyfor.users"], nil)
			return
		}
		err := session.RelationshipUserBlock(args[0])
		if err != nil {
			stdutil.PrintErr(lang["failed.block"], err)
			return
		}
	case "friends":
		if UserType != TypeUser {
			stdutil.PrintErr(lang["invalid.onlyfor.users"], nil)
			return
		}
		relations, err := session.RelationshipsGet()
		if err != nil {
			stdutil.PrintErr(lang["failed.friends"], err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Type", "Name")

		for _, relation := range relations {
			table.AddRow()
			table.AddStrings(relation.ID, TypeRelationships[relation.Type], relation.User.Username)
		}

		printTable(table)
	case "bookmarks":
		for key, _ := range bookmarks {
			fmt.Println(key)
		}
	case "bookmark":
		if nargs < 1 {
			stdutil.PrintErr("bookmark <name>", nil)
			return
		}

		key := strings.Join(args, " ")
		if strings.HasPrefix(key, "-") {
			key = key[1:]
			delete(bookmarks, key)
		} else {
			bookmarks[key] = loc.channel.ID
		}
		err := saveBookmarks()
		if err != nil {
			stdutil.PrintErr(lang["failed.file.save"], err)
		}
	case "go":
		if nargs < 1 {
			stdutil.PrintErr("go <bookmark>", nil)
			return
		}
		bookmark, ok := bookmarks[args[0]]
		if !ok {
			stdutil.PrintErr(lang["invalid.bookmark"], nil)
			return
		}

		var guild *discordgo.Guild
		var channel *discordgo.Channel
		var err error

		if bookmark != "" {
			channel, err = session.Channel(bookmark)
			if err != nil {
				stdutil.PrintErr(lang["failed.channel"], err)
				return
			}
		}

		if channel != nil && !channel.IsPrivate {
			guild, err = session.Guild(channel.GuildID)
			if err != nil {
				stdutil.PrintErr(lang["failed.guild"], err)
				return
			}
		}

		loc.push(guild, channel)
	case "tts":
		if nargs < 1 {
			stdutil.PrintErr("tts <stuff>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		msgStr := strings.Join(args, " ")
		if len(msgStr) > MsgLimit {
			stdutil.PrintErr(lang["invalid.limit.message"], nil)
			return
		}

		msg, err := session.ChannelMessageSendTTS(loc.channel.ID, msgStr)
		if err != nil {
			stdutil.PrintErr(lang["failed.msg.send"], err)
			return
		}
		fmt.Println(lang["status.msg.create"] + msg.ID)
		lastUsedMsg = msg.ID
		returnVal = msg.ID
	case "big":
		if nargs < 1 {
			stdutil.PrintErr("big <stuff>", nil)
			return
		}
		if loc.channel == nil && UserType != TypeWebhook {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		send := func(buffer string) (*discordgo.Message, bool) {
			if UserType == TypeWebhook {
				err := session.WebhookExecute(UserId, UserToken, false, &discordgo.WebhookParams{
					Content: buffer,
				})
				if err != nil {
					stdutil.PrintErr(lang["failed.msg.send"], err)
					return nil, false
				}
				return nil, true
			} else {
				msg, err := session.ChannelMessageSend(loc.channel.ID, buffer)
				if err != nil {
					stdutil.PrintErr(lang["failed.msg.send"], err)
					return nil, false
				}
				fmt.Println(lang["status.msg.create"] + msg.ID)

				return msg, true
			}
		}

		buffer := ""
		for _, c := range strings.Join(args, " ") {
			str := toEmojiString(c)
			if len(buffer)+len(str) > MsgLimit {
				_, ok := send(buffer)
				if !ok {
					return
				}

				buffer = ""
			}
			buffer += str
		}
		msg, _ := send(buffer)

		if msg != nil {
			lastUsedMsg = msg.ID
			returnVal = msg.ID
		}
	case "reactbig":
		if nargs < 2 {
			stdutil.PrintErr("reactbig <message id> <text>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		used := ""

		for _, c := range strings.Join(args[1:], " ") {
			str := string(toEmoji(c))

			if strings.Contains(used, str) {
				fmt.Println(lang["failed.react.used"])
				continue
			}
			used += str

			err := session.MessageReactionAdd(loc.channel.ID, args[0], str)
			if err != nil {
				stdutil.PrintErr(lang["failed.react"], err)
			}
		}
	case "ginfo":
		if nargs < 1 {
			stdutil.PrintErr("ginfo <property>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(lang["invalid.guild"], nil)
			return
		}

		switch strings.ToLower(args[0]) {
		case "name":
			returnVal = loc.guild.Name
		case "icon":
			returnVal = loc.guild.Icon
		case "region":
			returnVal = loc.guild.Region
		case "owner":
			returnVal = loc.guild.OwnerID
		case "splash":
			returnVal = loc.guild.Splash
		case "members":
			returnVal = strconv.Itoa(loc.guild.MemberCount)
		case "level":
			returnVal = TypeVerifications[loc.guild.VerificationLevel]
		default:
			stdutil.PrintErr(lang["invalid.value"], nil)
		}

		if returnVal != "" {
			fmt.Println(returnVal)
		}
	case "rl":
		full := nargs >= 1 && strings.EqualFold(args[0], "full")

		var err error
		if full {
			fmt.Println(lang["restarting.session"])
			err = session.Close()
			if err != nil {
				stdutil.PrintErr(lang["failed.session.close"], err)
				return
			}
			err = session.Open()
			if err != nil {
				stdutil.PrintErr(lang["failed.session.start"], err)
			}
		}

		fmt.Println(lang["restarting.cache.loc"])
		var guild *discordgo.Guild
		var channel *discordgo.Channel

		if loc.guild != nil {
			guild, err = session.Guild(loc.guild.ID)

			if err != nil {
				stdutil.PrintErr(lang["failed.guild"], err)
				return
			}
		}

		if loc.channel != nil {
			channel, err = session.Channel(loc.channel.ID)

			if err != nil {
				stdutil.PrintErr(lang["failed.channel"], err)
				return
			}
		}

		loc.guild = guild
		loc.channel = channel
		pointerCache = ""

		fmt.Println(lang["restarting.cache.vars"])
		cacheGuilds = make(map[string]string)
		cacheChannels = make(map[string]string)
		cacheAudio = make(map[string][][]byte)

		lastLoc = location{}
		lastMsg = location{}
		lastUsedMsg = ""
		lastUsedRole = ""

		cacheRead = nil
	case "uinfo":
		if nargs < 2 {
			stdutil.PrintErr("uinfo <user id> <property>", nil)
			return
		}
		id := args[0]

		if UserType != TypeBot && !strings.EqualFold(id, "@me") {
			stdutil.PrintErr(lang["invalid.onlyfor.bots"], nil)
			return
		}

		user, err := session.User(id)
		if err != nil {
			stdutil.PrintErr(lang["failed.user"], err)
			return
		}

		switch strings.ToLower(args[1]) {
		case "id":
			returnVal = user.ID
		case "email":
			returnVal = user.Email
		case "name":
			returnVal = user.Username
		case "avatar":
			returnVal = user.Avatar
		case "bot":
			returnVal = strconv.FormatBool(user.Bot)
		default:
			stdutil.PrintErr(lang["invalid.value"], nil)
		}

		if returnVal != "" {
			fmt.Println(returnVal)
		}
	case "avatar":
		if nargs < 1 {
			stdutil.PrintErr("avatar <file/link>", nil)
			return
		}

		var reader io.Reader
		resource := strings.Join(args, " ")

		if strings.HasPrefix(resource, "https://") || strings.HasPrefix(resource, "http://") {
			res, err := http.Get(resource)
			if err != nil {
				stdutil.PrintErr(lang["failed.webrequest"], err)
				return
			}
			defer res.Body.Close()

			reader = res.Body
		} else {
			err := fixPath(&resource)
			if err != nil {
				stdutil.PrintErr(lang["failed.fixpath"], err)
				return
			}

			r, err := os.Open(resource)
			defer r.Close()
			if err != nil {
				stdutil.PrintErr(lang["failed.file.open"], err)
				return
			}

			reader = r
		}

		writer := bytes.NewBuffer([]byte{})
		b64 := base64.NewEncoder(base64.StdEncoding, writer)

		_, err := io.Copy(b64, reader)
		if err != nil {
			stdutil.PrintErr(lang["failed.base64"], err)
			return
		}
		b64.Close()

		// Too lazy to detect image type. Seems to work anyway ¯\_(ツ)_/¯
		str := "data:image/png;base64," + writer.String()

		if UserType == TypeWebhook {
			_, err = session.WebhookEditWithToken(UserId, UserToken, "", str)
			if err != nil {
				stdutil.PrintErr(lang["failed.avatar"], err)
				return
			}
			return
		}

		user, err := session.User("@me")
		if err != nil {
			stdutil.PrintErr(lang["failed.user"], err)
			return
		}

		_, err = session.UserUpdate("", "", user.Username, str, "")
		if err != nil {
			stdutil.PrintErr(lang["failed.avatar"], err)
			return
		}
		fmt.Println(lang["status.avatar"])
	case "sayfile":
		if nargs < 1 {
			stdutil.PrintErr("sayfile <path>", nil)
			return
		}
		if loc.channel == nil && UserType != TypeWebhook {
			stdutil.PrintErr(lang["invalid.channel"], nil)
			return
		}

		path := args[0]
		err := fixPath(&path)
		if err != nil {
			stdutil.PrintErr(lang["failed.fixpath"], err)
			return
		}

		reader, err := os.Open(path)
		if err != nil {
			stdutil.PrintErr(lang["failed.file.open"], err)
			return
		}
		defer reader.Close()

		send := func(buffer string) (*discordgo.Message, bool) {
			if UserType == TypeWebhook {
				err = session.WebhookExecute(UserId, UserToken, false, &discordgo.WebhookParams{
					Content: buffer,
				})
				if err != nil {
					stdutil.PrintErr(lang["failed.msg.send"], err)
					return nil, false
				}
				return nil, true
			} else {
				msg, err := session.ChannelMessageSend(loc.channel.ID, buffer)
				if err != nil {
					stdutil.PrintErr(lang["failed.msg.send"], err)
					return nil, false
				}
				fmt.Println("Created message with ID " + msg.ID)

				return msg, true
			}
		}

		scanner := bufio.NewScanner(reader)
		buffer := ""

		for i := 1; scanner.Scan(); i++ {
			text := scanner.Text()
			if len(text) > MsgLimit {
				stdutil.PrintErr("Line "+strconv.Itoa(i)+" exceeded "+strconv.Itoa(MsgLimit)+" characters.", nil)
				return
			} else if len(buffer)+len(text) > MsgLimit {
				_, ok := send(buffer)
				if !ok {
					return
				}

				buffer = ""
			}
			buffer += text + "\n"
		}

		err = scanner.Err()
		if err != nil {
			stdutil.PrintErr(lang["failed.file.read"], err)
		}
		msg, _ := send(buffer)
		if msg != nil {
			returnVal = msg.ID
			lastUsedMsg = msg.ID
		}
	case "name":
		if nargs < 1 {
			stdutil.PrintErr("name <handle>", nil)
			return
		}

		if UserType == TypeWebhook {
			_, err := session.WebhookEditWithToken(UserId, UserToken, strings.Join(args, " "), "")
			if err != nil {
				stdutil.PrintErr(lang["failed.user.edit"], err)
			}
			return
		}

		user, err := session.User("@me")
		if err != nil {
			stdutil.PrintErr(lang["failed.user"], err)
			return
		}

		user, err = session.UserUpdate("", "", strings.Join(args, " "), user.Avatar, "")
		if err != nil {
			stdutil.PrintErr(lang["failed.user.edit"], err)
			return
		}
		fmt.Println(lang["status.name"])
	case "status":
		if nargs < 1 {
			stdutil.PrintErr("status <value>", nil)
			return
		}
		status, ok := TypeStatuses[strings.ToLower(args[0])]
		if !ok {
			stdutil.PrintErr(lang["invalid.value"], nil)
			return
		}

		if status == discordgo.StatusOffline {
			stdutil.PrintErr(lang["invalid.status.offline"], nil)
			return
		}

		_, err := session.UserUpdateStatus(status)
		if err != nil {
			stdutil.PrintErr(lang["failed.status"], err)
			return
		}
		fmt.Println(lang["status.status"])
	default:
		stdutil.PrintErr(lang["invalid.command"], nil)
	}
	return
}

func channels(session *discordgo.Session, kind string) {
	if loc.guild == nil {
		stdutil.PrintErr(lang["invalid.guild"], nil)
		return
	}
	channels, err := session.GuildChannels(loc.guild.ID)
	if err != nil {
		stdutil.PrintErr(lang["failed.channel"], nil)
		return
	}

	cacheChannels = make(map[string]string)

	sort.Slice(channels, func(i int, j int) bool {
		return channels[i].Position < channels[j].Position
	})

	table := gtable.NewStringTable()
	table.AddStrings("ID", "Name")

	for _, channel := range channels {
		if channel.Type != kind {
			continue
		}
		table.AddRow()
		table.AddStrings(channel.ID, channel.Name)
		cacheChannels[strings.ToLower(channel.Name)] = channel.ID
	}

	printTable(table)
}

func parseBool(str string) (bool, error) {
	if str == "yes" || str == "true" || str == "y" {
		return true, nil
	} else if str == "no" || str == "false" || str == "n" {
		return false, nil
	}
	return false, errors.New(lang["invalid.yn"])
}

func printTable(table gtable.StringTable) {
	table.Each(func(ti *gtable.TableItem) {
		ti.Padding(1)
	})
	fmt.Println(table.String())
}
