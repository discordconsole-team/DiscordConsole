package main

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/gtable"
	"github.com/legolord208/stdutil"
)

var typeRelationships = map[int]string{
	1: "Friend",
	2: "Blocked",
	3: "Incoming request",
	4: "Sent request",
}
var typeVerifications = map[discordgo.VerificationLevel]string{
	discordgo.VerificationLevelNone:   "None",
	discordgo.VerificationLevelLow:    "Low",
	discordgo.VerificationLevelMedium: "Medium",
	discordgo.VerificationLevelHigh:   "High",
}
var typeMessages = map[string]int{
	"all":      messagesAll,
	"mentions": messagesMentions,
	"private":  messagesPrivate,
	"current":  messagesCurrent,
	"none":     messagesNone,
}
var typeStatuses = map[string]discordgo.Status{
	"online":    discordgo.StatusOnline,
	"idle":      discordgo.StatusIdle,
	"dnd":       discordgo.StatusDoNotDisturb,
	"invisible": discordgo.StatusInvisible,
}

type location struct {
	guild   *discordgo.Guild
	channel *discordgo.Channel
}

func (loc *location) push(guild *discordgo.Guild, channel *discordgo.Channel) {
	sameGuild := guild == loc.guild || (loc.guild != nil && guild != nil && loc.guild.ID == guild.ID)
	sameChannel := channel == loc.channel || (loc.channel != nil && channel != nil && loc.channel.ID == channel.ID)

	if sameGuild && sameChannel {
		return
	}

	lastLoc = *loc

	loc.guild = guild
	loc.channel = channel
	pointerCache = ""

	if !sameGuild {
		cacheGuilds = nil
		cacheChannels = nil
	}
}

var loc location
var lastLoc location
var lastMsg location

var lastUsedMsg string
var lastUsedRole string

var cacheGuilds []*discordgo.UserGuild
var cacheChannels []*discordgo.Channel
var cachedChannelType string

var cacheRead *discordgo.Message
var cacheUser *discordgo.User

const (
	messagesNone = iota
	messagesCurrent
	messagesPrivate
	messagesMentions
	messagesAll
)

var messages = messagesNone
var intercept = true
var output = false

var webhookCommands = []string{"big", "say", "sayfile", "embed", "name", "avatar", "exit", "exec", "run"}

func command(session *discordgo.Session, cmd string, w io.Writer) (returnVal string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	parts := strings.FieldsFunc(cmd, func(c rune) bool {
		return c != '\n' && unicode.IsSpace(c)
	})

	cmd = strings.ToLower(parts[0])
	args := parts[1:]

	returnVal = commandRaw(session, cmd, args, w)
	return
}

func commandRaw(session *discordgo.Session, cmd string, args []string, w io.Writer) (returnVal string) {
	defer handleCrash()
	nargs := len(args)

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
		exit(session)
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

		err = runLua(session, script, scriptArgs...)
		if err != nil {
			stdutil.PrintErr(tl("failed.lua.run"), err)
		}
	case "guilds":
		fallthrough
	case "guild":
		fallthrough
	case "channels":
		fallthrough
	case "channel":
		fallthrough
	case "dm":
		fallthrough
	case "pchannels":
		fallthrough
	case "bookmarks":
		fallthrough
	case "bookmark":
		fallthrough
	case "go":
		returnVal = commandsNavigate(session, cmd, args, nargs, w)
	case "say":
		fallthrough
	case "tts":
		fallthrough
	case "embed":
		fallthrough
	case "quote":
		fallthrough
	case "big":
		fallthrough
	case "file":
		fallthrough
	case "edit":
		fallthrough
	case "editembed":
		fallthrough
	case "sayfile":
		returnVal = commandsSay(session, cmd, args, nargs, w)
	case "log":
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		directly := nargs < 1

		var file io.Writer

		if directly {
			file = w
		} else {
			name := strings.Join(args, " ")
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
		}

		limit := 100
		if directly {
			limit = 10
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
			err = writeln(file, s+msg.Author.Username+": "+msg.Content)
			if err != nil && !directly {
				stdutil.PrintErr(tl("failed.msg.write"), err)
				return
			}
		}
	case "delall":
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		since := ""
		if nargs >= 1 {
			since = args[0]
		}
		messages, err := session.ChannelMessages(loc.channel.ID, 100, "", since, "")
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.query"), err)
			return
		}

		ids := make([]string, len(messages))
		for i, msg := range messages {
			ids[i] = msg.ID
		}

		err = session.ChannelMessagesBulkDelete(loc.channel.ID, ids)
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.query"), err)
			return
		}
		returnVal := strconv.Itoa(len(ids))
		writeln(w, strings.Replace(tl("status.msg.delall"), "#", returnVal, -1))
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
			stdutil.PrintErr("invite accept <code> OR invite create [expire] [max uses] ['temp']", nil)
			return
		}
		switch args[0] {
		case "see":
			if nargs < 2 {
				stdutil.PrintErr("invite see <code>", nil)
				return
			}

			invite, err := session.Invite(args[1])
			if err != nil {
				stdutil.PrintErr(tl("failed.invite"), err)
				return
			}
			writeln(w, "Guild: "+invite.Guild.ID+", "+invite.Guild.Name)
			writeln(w, "Channel: "+invite.Channel.ID+", "+invite.Channel.Name)
			writeln(w, "Created At: "+string(invite.CreatedAt))
			writeln(w, "Inviter: "+invite.Inviter.ID+", "+invite.Inviter.String())
			writeln(w, "Max age: "+(time.Duration(invite.MaxAge)*time.Second).String())
			writeln(w, "Max uses: "+strconv.Itoa(invite.MaxUses))
			writeln(w, "Uses: "+strconv.Itoa(invite.Uses))
			writeln(w, "Revoked: "+strconv.FormatBool(invite.Revoked))
			writeln(w, "Temporary: "+strconv.FormatBool(invite.Temporary))
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
	case "reply":
		loc.push(lastMsg.guild, lastMsg.channel)
	case "back":
		loc.push(lastLoc.guild, lastLoc.channel)
	case "ban":
		if nargs < 1 {
			stdutil.PrintErr("ban <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildBanCreate(loc.guild.ID, args[0], 0)
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
			stdutil.PrintErr("kick <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildMemberDelete(loc.guild.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.kick"), err)
		}
	case "leave":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildLeave(loc.guild.ID)
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

		bans, err := session.GuildBans(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.ban.list"), err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("User ID", "Username", "Reason")

		for _, ban := range bans {
			table.AddRow()
			table.AddStrings(ban.User.ID, ban.User.Username, ban.Reason)
		}

		writeln(w, table.String())
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
	case "vchannels":
		channels(session, "voice", w)
	case "play":
		if userType != typeBot {
			stdutil.PrintErr(tl("invalid.onlyfor.bots"), nil)
			return
		}
		if nargs < 1 {
			stdutil.PrintErr("play <dca audio file>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
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
	case "reactadd":
		fallthrough
	case "reactdel":
		if nargs < 2 {
			stdutil.PrintErr("reactadd/reactdel <message id> <emoji unicode/id>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		var err error
		if cmd == "reactadd" {
			err = session.MessageReactionAdd(loc.channel.ID, args[0], args[1])
		} else {
			err = session.MessageReactionRemove(loc.channel.ID, args[0], args[1], "@me")
		}
		if err != nil {
			stdutil.PrintErr(tl("failed.react"), err)
			return
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
	case "friends":
		if userType != typeUser {
			stdutil.PrintErr(tl("invalid.onlyfor.users"), nil)
			return
		}
		relations, err := session.RelationshipsGet()
		if err != nil {
			stdutil.PrintErr(tl("failed.friends"), err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Type", "Name")

		for _, relation := range relations {
			table.AddRow()
			table.AddStrings(relation.ID, typeRelationships[relation.Type], relation.User.Username)
		}

		writeln(w, table.String())
	case "reactbig":
		if nargs < 2 {
			stdutil.PrintErr("reactbig <message id> <text>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		used := ""

		for _, c := range strings.Join(args[1:], " ") {
			str := string(toEmoji(c))

			if strings.Contains(used, str) {
				writeln(w, tl("failed.react.used"))
				continue
			}
			used += str

			err := session.MessageReactionAdd(loc.channel.ID, args[0], str)
			if err != nil {
				stdutil.PrintErr(tl("failed.react"), err)
			}
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

		lastLoc = location{}
		lastMsg = location{}
		lastUsedMsg = ""
		lastUsedRole = ""

		cacheRead = nil
		cacheUser = nil
	case "status":
		if nargs < 1 {
			stdutil.PrintErr("status <value>", nil)
			return
		}
		status, ok := typeStatuses[strings.ToLower(args[0])]
		if !ok {
			stdutil.PrintErr(tl("invalid.value"), nil)
			return
		}

		if status == discordgo.StatusOffline {
			stdutil.PrintErr(tl("invalid.status.offline"), nil)
			return
		}

		_, err := session.UserUpdateStatus(status)
		if err != nil {
			stdutil.PrintErr(tl("failed.status"), err)
			return
		}
		writeln(w, tl("status.status"))
	case "avatar":
		fallthrough
	case "name":
		fallthrough
	case "playing":
		fallthrough
	case "streaming":
		fallthrough
	case "typing":
		fallthrough
	case "nick":
		returnVal = commandsUserMod(session, cmd, args, nargs, w)
	case "read":
		fallthrough
	case "cinfo":
		fallthrough
	case "ginfo":
		fallthrough
	case "uinfo":
		returnVal = commandsQuery(session, cmd, args, nargs, w)
	case "roles":
		fallthrough
	case "roleadd":
		fallthrough
	case "roledel":
		fallthrough
	case "rolecreate":
		fallthrough
	case "roleedit":
		fallthrough
	case "roledelete":
		returnVal = commandsRoles(session, cmd, args, nargs, w)
	case "api_start":
		if apiName != "" {
			stdutil.PrintErr(tl("invalid.api.started"), nil)
			return
		}

		var name string
		if nargs >= 1 {
			name = strings.Join(args, " ")
			go apiStartName(session, name)
		} else {
			var err error
			name, err = apiStart(session)
			if err != nil {
				stdutil.PrintErr(tl("failed.api.start"), err)
				return
			}
		}
		writeln(w, tl("status.api.start")+" "+name)
		returnVal = name
	case "broadcast":
		if nargs < 1 {
			stdutil.PrintErr("broadcast <command>", nil)
			return
		}
		if apiName == "" {
			stdutil.PrintErr(tl("invalid.api.notstarted"), nil)
			return
		}

		err := apiSend(strings.Join(args, " "))
		if err != nil {
			stdutil.PrintErr(tl("failed.generic"), err)
			return
		}
		commandRaw(session, args[0], args[1:], w)
	case "api_stop":
		apiStop()
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
	default:
		stdutil.PrintErr(tl("invalid.command"), nil)
	}
	return
}

func parseBool(str string) (bool, error) {
	if str == "yes" || str == "true" || str == "y" {
		return true, nil
	} else if str == "no" || str == "false" || str == "n" {
		return false, nil
	}
	return false, errors.New(tl("invalid.yn"))
}
