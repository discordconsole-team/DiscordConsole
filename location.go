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
	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/stdutil"
)

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
		cacheChannels = nil
	}

	var err error
	if vc != nil {
		playing = ""
		if channel != nil && channel.Type == typeChannelVoice {
			err = vc.ChangeChannel(channel.ID, false, false)
		} else {
			err = vc.Disconnect()
			vc = nil
		}
		if err != nil {
			stdutil.PrintErr(tl("failed.voice.disconnect"), err)
		}
	} else if guild != nil && channel != nil && channel.Type == typeChannelVoice {
		vc, err = session.ChannelVoiceJoin(guild.ID, channel.ID, false, false)
		if err != nil {
			stdutil.PrintErr(tl("failed.voice.connect"), err)
		}
	}
}

var loc location
var lastLoc location
var lastMsg location
