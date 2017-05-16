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
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/legolord208/stdutil"
)

func messageCreate(session *discordgo.Session, e *discordgo.MessageCreate) {
	defer handleCrash()

	if e.Author == nil {
		return
	}

	var channel *discordgo.Channel
	var err error
	for _, c := range cacheChannels {
		if c.ID == e.ChannelID {
			channel = c
			break
		}
	}

	if channel == nil {
		channel, err = session.Channel(e.ChannelID)
		if err != nil {
			stdutil.PrintErr(tl("failed.channel"), err)
			return
		}
	}

	var guild *discordgo.Guild
	if !channel.IsPrivate {
		// Can't use cache. It's of user guild
		guild, err = session.Guild(channel.GuildID)
		if err != nil {
			stdutil.PrintErr(tl("failed.guild"), err)
			return
		}
	}

	if messageCommand(session, e.Message, guild, channel) {
		return
	}

	hasOutput := false

	print := false
outer:
	switch messages {
	case messagesAll:
		print = true
	case messagesPrivate:
		if channel.IsPrivate {
			print = true
		}
	case messagesMentions:
		if channel.IsPrivate || e.MentionEveryone {
			print = true
			break
		}

		for _, u := range e.Mentions {
			if u.ID == userID {
				print = true
				break outer
			}
		}

		user, err := session.State.Member(guild.ID, userID)
		if err != nil {
			stdutil.PrintErr(tl("failed.user"), err)
			break
		}

		for _, role := range user.Roles {
			for _, role2 := range e.MentionRoles {
				if role == role2 {
					print = true
					break outer
				}
			}
		}
	case messagesCurrent:
		if (guild == nil || loc.guild == nil) && loc.channel != nil && channel.ID != loc.channel.ID {
			break
		}
		if guild != nil && loc.guild != nil && guild.ID != loc.guild.ID {
			break
		}

		print = true
	}
	if print {
		printMessage(session, e.Message, true, guild, channel, color.Output)
		hasOutput = true
	}

	if len(luaMessageEvents) > 0 {
		hasOutput = true

		color.Unset()
		colorAutomated.Set()

		fmt.Print("\r" + strings.Repeat(" ", 20) + "\r")
		luaMessageEvent(session, e.Message)

		color.Unset()
		colorDefault.Set()
	}
	if hasOutput {
		printPointer(session)
	}
}

func messageCommand(session *discordgo.Session, e *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel) (isCmd bool) {
	if e.Author.ID != userID {
		return
	} else if !intercept {
		return
	}

	prefix := tl("console.")

	contents := strings.TrimSpace(e.Content)
	if !strings.HasPrefix(contents, prefix) {
		return
	}
	cmd := contents[len(prefix):]

	isCmd = true

	if strings.EqualFold(cmd, "ping") {
		first := time.Now().UTC()

		timestamp, err := e.Timestamp.Parse()
		if err != nil {
			stdutil.PrintErr(tl("failed.timestamp"), err)
			return
		}

		in := first.Sub(timestamp)

		// Discord 'bug' makes us receive the message before the timestamp, sometimes.
		text := ""
		if in.Nanoseconds() >= 0 {
			text += "Incoming: `" + in.String() + "`"
		} else {
			text += "Message was received earlier than timestamp. Discord bug."
		}

		middle := time.Now().UTC()

		_, err = session.ChannelMessageEditComplex(discordgo.NewMessageEdit(e.ChannelID, e.ID).
			SetContent("Pong! 1/2").
			SetEmbed(&discordgo.MessageEmbed{
				Description: text + "\nCalculating outgoing..",
			}))

		last := time.Now().UTC()

		text += "\nOutgoing: `" + last.Sub(middle).String() + "`"
		text += "\n\n\nIncoming is the time it takes for the message to reach DiscordConsole."
		text += "\nOutgoing is the time it takes for DiscordConsole to reach discord."

		_, err = session.ChannelMessageEditComplex(discordgo.NewMessageEdit(e.ChannelID, e.ID).
			SetContent("Pong! 2/2").
			SetEmbed(&discordgo.MessageEmbed{
				Description: text,
			}))
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.edit"), err)
		}
		return
	}

	loc.push(guild, channel)

	capture := output

	var w io.Writer
	var str *bytes.Buffer
	if capture {
		str = bytes.NewBuffer(nil)
		w = str
	} else {
		go func() {
			err := session.ChannelMessageDelete(e.ChannelID, e.ID)
			if err != nil {
				stdutil.PrintErr(tl("failed.msg.delete"), err)
			}
		}()
		color.Unset()
		colorAutomated.Set()

		fmt.Println(cmd)
		w = color.Output
	}
	command(session, commandSource{}, cmd, w)

	if !capture {
		color.Unset()
		colorDefault.Set()
		printPointer(session)
	} else {
		first := true
		send := func(buf string) {
			if buf == "" {
				return
			}

			// Zero width space
			buf = "```\n" + strings.Replace(buf, "`", "​`", -1) + "\n```"
			if first {
				first = false
				_, err := session.ChannelMessageEdit(e.ChannelID, e.ID, buf)
				if err != nil {
					stdutil.PrintErr(tl("failed.msg.edit"), err)
					return
				}
			} else {
				_, err := session.ChannelMessageSend(e.ChannelID, buf)
				if err != nil {
					stdutil.PrintErr(tl("failed.msg.send"), err)
					return
				}
			}
		}

		buf := ""
		for {
			line, err := str.ReadString('\n')
			if err != nil {
				break
			}

			if len(line)+len(buf)+8 < msgLimit {
				buf += line
			} else {
				send(buf)
				buf = ""
			}
		}
		send(buf)
	}

	color.Unset()
	colorDefault.Set()
	return
}
