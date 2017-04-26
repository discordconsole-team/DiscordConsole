package main

import "github.com/bwmarrin/discordgo"

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
}

var loc location
var lastLoc location
var lastMsg location
