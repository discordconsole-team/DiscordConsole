/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2017  LEGOlord208

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
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/stdutil"
)

func commandsQuery(session *discordgo.Session, cmd string, args []string, nargs int, w io.Writer) (returnVal string) {
	switch cmd {
	case "read":
		if nargs < 1 {
			stdutil.PrintErr("read <message id> [property]", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		msgID := args[0]

		var msg *discordgo.Message
		var err error
		if strings.EqualFold(msgID, "cache") {
			if cacheRead == nil {
				stdutil.PrintErr(tl("invalid.cache"), nil)
				return
			}

			msg = cacheRead
		} else {
			msg, err = getMessage(session, loc.channel.ID, msgID)
			if err != nil {
				stdutil.PrintErr(tl("failed.msg.query"), err)
				return
			}

			cacheRead = msg
		}

		property := ""
		if len(args) >= 2 {
			property = strings.ToLower(args[1])
		}
		switch property {
		case "":
			printMessage(session, msg, false, loc.guild, loc.channel, w)
		case "text":
			returnVal = msg.Content
		case "channel":
			returnVal = msg.ChannelID
		case "timestamp":
			t, err := timestamp(msg)
			if err != nil {
				stdutil.PrintErr(tl("failed.timestamp"), err)
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
		case "embed":
			embed := "{}"
			if len(msg.Embeds) >= 1 {
				embedBytes, err := json.MarshalIndent(msg.Embeds[0], "", "\t")
				if err != nil {
					stdutil.PrintErr("Failed to make embed into JSON", err)
					return
				}
				embed = string(embedBytes)
			}
			returnVal = embed
		default:
			stdutil.PrintErr(tl("invalid.value"), nil)
		}

		lastUsedMsg = msg.ID
		if returnVal != "" {
			writeln(w, returnVal)
		}
	case "cinfo":
		if nargs < 1 {
			stdutil.PrintErr("cinfo <property>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
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
			stdutil.PrintErr(tl("invalid.value"), nil)
		}

		if returnVal != "" {
			writeln(w, returnVal)
		}
	case "ginfo":
		if nargs < 1 {
			stdutil.PrintErr("ginfo <property>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
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
			returnVal = typeVerifications[loc.guild.VerificationLevel]
		default:
			stdutil.PrintErr(tl("invalid.value"), nil)
		}

		if returnVal != "" {
			writeln(w, returnVal)
		}
	case "uinfo":
		if nargs < 2 {
			stdutil.PrintErr("uinfo <user id> <property>", nil)
			return
		}
		id := args[0]
		var user *discordgo.User

		if strings.EqualFold(id, "cache") {
			if cacheUser == nil {
				stdutil.PrintErr(tl("invalid.cache"), nil)
				return
			}

			user = cacheUser
		} else {
			if userType != typeBot && !strings.EqualFold(id, "@me") {
				stdutil.PrintErr(tl("invalid.onlyfor.bots"), nil)
				return
			}

			var err error
			user, err = session.User(id)
			if err != nil {
				stdutil.PrintErr(tl("failed.user"), err)
				return
			}

			cacheUser = user
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
			stdutil.PrintErr(tl("invalid.value"), nil)
		}

		if returnVal != "" {
			writeln(w, returnVal)
		}
	}
	return
}
