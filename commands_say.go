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
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/jD91mZM2/stdutil"
)

func commandsSay(session *discordgo.Session, source commandSource, cmd string, args []string, nargs int, w io.Writer) (returnVal string) {
	switch cmd {
	case "tts":
		fallthrough
	case "say":
		if nargs < 1 {
			stdutil.PrintErr("say/tts <stuff>", nil)
			return
		}
		if loc.channel == nil && userType != typeWebhook {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		toggle := false
		tts := cmd == "tts"
		parts := args

		sendbuf := func(buffer string) (ok bool) {
			if userType == typeWebhook {
				err := session.WebhookExecute(userID, userToken, false, &discordgo.WebhookParams{
					Content: buffer,
					TTS:     tts,
				})
				if err != nil {
					stdutil.PrintErr(tl("failed.msg.send"), err)
					return
				}
				ok = true
				return
			}

			msg, err := session.ChannelMessageSendComplex(loc.channel.ID, &discordgo.MessageSend{
				Content: buffer,
				Tts:     tts,
			})
			if err != nil {
				stdutil.PrintErr(tl("failed.msg.send"), err)
				return
			}
			writeln(w, tl("status.msg.create")+" "+msg.ID)
			lastUsedMsg = msg.ID
			returnVal = msg.ID
			ok = true
			return
		}

	outer:
		for {
			msgStr := strings.Join(parts, " ")
			if source.Terminal && msgStr == "toggle" {
				toggle = !toggle
			} else {
				/*
					if len(msgStr) > msgLimit {
						stdutil.PrintErr(tl("invalid.limit.message"), nil)
						return
					}
				*/

				for len(msgStr) > msgLimit {
					buffer := msgStr[:msgLimit]
					msgStr = msgStr[msgLimit:]

					if !sendbuf(buffer) {
						return
					}
				}
				sendbuf(msgStr)
			}

			if !toggle {
				break
			}

			for {
				color.Unset()
				colorChatMode.Set()

				text, err := rl.Readline()
				if err != nil {
					if err != readline.ErrInterrupt && err != io.EOF {
						stdutil.PrintErr(tl("failed.readline.read"), err)
					}
					break outer
				}

				color.Unset()

				parts, err = parse(substitute, text)
				if len(parts) >= 1 {
					break
				}
			}
		}
	case "embed":
		if nargs < 1 {
			stdutil.PrintErr("embed <embed json>", nil)
			return
		}
		if loc.channel == nil && userType != typeWebhook {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		jsonstr := strings.Join(args, " ")
		var embed = &discordgo.MessageEmbed{}

		err := json.Unmarshal([]byte(jsonstr), embed)
		if err != nil {
			stdutil.PrintErr(tl("failed.json"), err)
			return
		}

		if userType == typeWebhook {
			err = session.WebhookExecute(userID, userToken, false, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{embed},
			})
			if err != nil {
				stdutil.PrintErr(tl("failed.msg.send"), err)
				return
			}
		} else {
			msg, err := session.ChannelMessageSendEmbed(loc.channel.ID, embed)
			if err != nil {
				stdutil.PrintErr(tl("failed.msg.send"), err)
				return
			}
			writeln(w, tl("status.msg.create")+" "+msg.ID)
			lastUsedMsg = msg.ID
			returnVal = msg.ID
		}
	case "big":
		if nargs < 1 {
			stdutil.PrintErr("big <stuff>", nil)
			return
		}
		if loc.channel == nil && userType != typeWebhook {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		buffer := ""
		for _, c := range strings.Join(args, " ") {
			str := toEmojiString(c)
			if len(buffer)+len(str) > msgLimit {
				var msg *discordgo.Message
				if userType == typeWebhook { // Webhook checking.
					msg, _ = say(session, w, "", buffer)
				} else {
					msg, _ = say(session, w, loc.channel.ID, buffer)
				}

				if msg != nil {
					lastUsedMsg = msg.ID
					returnVal = msg.ID
				}

				buffer = ""
			}
			buffer += str
		}
		if buffer != "" {
			var msg *discordgo.Message
			if userType == typeWebhook { // Webhook checking.
				msg, _ = say(session, w, "", buffer)
			} else {
				msg, _ = say(session, w, loc.channel.ID, buffer)
			}

			if msg != nil {
				lastUsedMsg = msg.ID
				returnVal = msg.ID
			}
		}
	case "sayfile":
		if nargs < 1 {
			stdutil.PrintErr("sayfile <path>", nil)
			return
		}
		if loc.channel == nil && userType != typeWebhook {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		channel := ""
		if loc.channel != nil {
			channel = loc.channel.ID
		}

		path := args[0]
		err := fixPath(&path)
		if err != nil {
			stdutil.PrintErr(tl("failed.fixpath"), err)
			return
		}

		reader, err := os.Open(path)
		if err != nil {
			stdutil.PrintErr(tl("failed.file.open"), err)
			return
		}
		defer reader.Close()

		scanner := bufio.NewScanner(reader)
		buffer := ""

		for i := 1; scanner.Scan(); i++ {
			text := scanner.Text()
			if len(text) > msgLimit {
				stdutil.PrintErr("Line "+strconv.Itoa(i)+" exceeded "+strconv.Itoa(msgLimit)+" characters.", nil)
				return
			} else if len(buffer)+len(text) > msgLimit {
				msg, ok := say(session, w, channel, buffer)
				if !ok {
					return
				}
				if msg != nil {
					returnVal = msg.ID
					lastUsedMsg = msg.ID
				}

				buffer = ""
			}
			buffer += text + "\n"
		}

		err = scanner.Err()
		if err != nil {
			stdutil.PrintErr(tl("failed.file.read"), err)
			return
		}
		if buffer != "" {
			msg, _ := say(session, w, channel, buffer)
			if msg != nil {
				returnVal = msg.ID
				lastUsedMsg = msg.ID
			}
		}
	case "file":
		if nargs < 1 {
			stdutil.PrintErr("file <file>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		name := strings.Join(args, " ")
		err := fixPath(&name)
		if err != nil {
			stdutil.PrintErr(tl("failed.fixpath"), err)
		}

		file, err := os.Open(name)
		if err != nil {
			stdutil.PrintErr(tl("failed.file.open"), nil)
			return
		}
		msg, err := session.ChannelFileSend(loc.channel.ID, filepath.Base(name), file)
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.send"), err)
			return
		}
		file.Close()
		writeln(w, tl("status.msg.create")+" "+msg.ID)
		returnVal = msg.ID
	case "quote":
		if nargs < 1 {
			stdutil.PrintErr("quote <message id>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		msg, err := getMessage(session, loc.channel.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.query"), err)
			return
		}

		t, err := timestamp(msg)
		if err != nil {
			stdutil.PrintErr(tl("failed.timestamp"), err)
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
			stdutil.PrintErr(tl("failed.msg.send"), err)
			return
		}
		writeln(w, tl("status.msg.create")+" "+msg.ID)
		lastUsedMsg = msg.ID
		returnVal = msg.ID
	case "editembed":
		fallthrough
	case "edit":
		if nargs < 2 {
			stdutil.PrintErr("edit <message id> <stuff>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		id := args[0]
		contents := strings.Join(args[1:], " ")

		var msg *discordgo.Message
		var err error
		if cmd == "editembed" {
			var embed = &discordgo.MessageEmbed{}
			err := json.Unmarshal([]byte(contents), embed)
			if err != nil {
				stdutil.PrintErr(tl("failed.json"), err)
				return
			}

			msg, err = session.ChannelMessageEditEmbed(loc.channel.ID, id, embed)
		} else {
			msg, err = session.ChannelMessageEdit(loc.channel.ID, id, contents)
		}
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.edit"), err)
			return
		}
		lastUsedMsg = msg.ID
	case "del":
		if nargs < 1 {
			stdutil.PrintErr("del <message id>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		err := session.ChannelMessageDelete(loc.channel.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.delete"), err)
			return
		}
	case "delall":
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		if isPrivate(loc.channel) {
			stdutil.PrintErr(tl("invalid.dm"), nil)
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
	}
	return
}
