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
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jD91mZM2/stdutil"
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
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		values := chan2array(loc.channel)

		if nargs < 1 {
			for _, keyval := range values {
				writeln(w, keyval.String())
			}
		} else {
			var ok bool
			returnVal, ok = findValByKey(values, args[0])
			if !ok {
				stdutil.PrintErr(tl("invalid.value"), nil)
				return
			}

			writeln(w, returnVal)
		}
	case "ginfo":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		values := guild2array(loc.guild)

		if nargs < 1 {
			for _, keyval := range values {
				writeln(w, keyval.String())
			}
		} else {
			var ok bool
			returnVal, ok = findValByKey(values, args[0])
			if !ok {
				stdutil.PrintErr(tl("invalid.value"), nil)
				return
			}

			writeln(w, returnVal)
		}
	case "uinfo":
		if nargs < 1 {
			stdutil.PrintErr("uinfo <id> [property]", nil)
			return
		}
		id := args[0]
		var keyvals []*keyval

		if strings.EqualFold(id, "cache") {
			if cacheUser == nil {
				stdutil.PrintErr(tl("invalid.cache"), nil)
				return
			}

			keyvals = cacheUser
		} else {

			user, err := session.User(id)
			if err != nil {
				stdutil.PrintErr(tl("failed.user"), err)
				return
			}

			keyvals = user2array(user)
			cacheUser = keyvals
		}

		if nargs < 2 {
			for _, keyval := range keyvals {
				writeln(w, keyval.String())
			}
		} else {
			var ok bool
			returnVal, ok = findValByKey(keyvals, args[1])
			if !ok {
				stdutil.PrintErr(tl("invalid.value"), nil)
				return
			}

			writeln(w, returnVal)
		}
	}
	return
}

func guild2array(guild *discordgo.Guild) []*keyval {
	return []*keyval{
		&keyval{"ID", guild.ID},
		&keyval{"Name", guild.Name},
		&keyval{"Icon", guild.Icon},
		&keyval{"Region", guild.Region},
		&keyval{"Owner", guild.OwnerID},
		&keyval{"Join messages", guild.SystemChannelID},
		&keyval{"Widget channel", guild.WidgetChannelID},
		&keyval{"AFK channel", guild.AfkChannelID},
		&keyval{"AFK timeout", strconv.Itoa(guild.AfkTimeout)},
		&keyval{"Members", strconv.Itoa(guild.MemberCount)},
		&keyval{"Verification", typeVerifications[guild.VerificationLevel]},
		&keyval{"Admin MFA", typeMfa[guild.MfaLevel]},
		&keyval{"Explicit Content Filter", typeContentFilter[guild.ExplicitContentFilter]},
		&keyval{"Unavailable", strconv.FormatBool(guild.Unavailable)},
	}
}

func chan2array(channel *discordgo.Channel) []*keyval {
	return []*keyval{
		&keyval{"ID", channel.ID},
		&keyval{"Guild", channel.GuildID},
		&keyval{"Name", channel.Name},
		&keyval{"Topic", channel.Topic},
		&keyval{"Type", typeChannel[channel.Type]},
		&keyval{"NSFW", strconv.FormatBool(channel.NSFW)},
		&keyval{"Parent category", channel.ParentID},
		&keyval{"Last message", channel.LastMessageID},
		&keyval{"Bitrate", strconv.Itoa(channel.Bitrate)},
		&keyval{"User limit", strconv.Itoa(channel.UserLimit)},
	}
}

func user2array(user *discordgo.User) []*keyval {
	return []*keyval{
		&keyval{"ID", user.ID},
		&keyval{"Email", user.Email},
		&keyval{"Name", user.Username},
		&keyval{"Discrim", user.Discriminator},
		&keyval{"Locale", user.Locale},
		&keyval{"Avatar", user.Avatar},
		&keyval{"Avatar URL", user.AvatarURL("1024")},
		&keyval{"Verified", strconv.FormatBool(user.Verified)},
		&keyval{"MFA Enabled", strconv.FormatBool(user.MFAEnabled)},
		&keyval{"Bot", strconv.FormatBool(user.Bot)},
	}
}

func invite2array(invite *discordgo.Invite) []*keyval {
	values := make([]*keyval, 0)
	if invite.Inviter != nil {
		for _, keyval := range user2array(invite.Inviter) {
			keyval.Key = "Inviter_" + keyval.Key
			values = append(values, keyval)
		}
	}
	for _, keyval := range guild2array(invite.Guild) {
		keyval.Key = "Guild_" + keyval.Key
		values = append(values, keyval)
	}
	for _, keyval := range chan2array(invite.Channel) {
		keyval.Key = "Channel_" + keyval.Key
		values = append(values, keyval)
	}
	values = append(values,
		&keyval{"Created_at", string(invite.CreatedAt)},
		&keyval{"Max_age", (time.Duration(invite.MaxAge) * time.Second).String()},
		&keyval{"Max_uses", strconv.Itoa(invite.MaxUses)},
		&keyval{"Uses", strconv.Itoa(invite.Uses)},
		&keyval{"Revoked", strconv.FormatBool(invite.Revoked)},
		&keyval{"Temporary", strconv.FormatBool(invite.Temporary)},
	)
	return values
}
