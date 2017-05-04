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

import "github.com/bwmarrin/discordgo"

var cacheGuilds []*discordgo.UserGuild
var cacheChannels []*discordgo.Channel
var cachedChannelType string

var chanReady = make(chan []*discordgo.UserGuild)

func ready(session *discordgo.Session, e *discordgo.Ready) {
	select {
	case _, ok := <-chanReady:
		if !ok {
			return
		}
	default:
	}

	uguilds := make([]*discordgo.UserGuild, len(e.Guilds))
	for i, guild := range e.Guilds {
		uguilds[i] = toUserGuild(guild)
	}
	guilds := sortGuilds(uguilds, e.Settings)

	cacheGuilds = guilds

	select {
	case chanReady <- guilds:
	default:
	}
	close(chanReady)
}

func guildCreate(session *discordgo.Session, e *discordgo.GuildCreate) {
	// Fot bots, the guildcreate event triggers on startup.
	for _, guild := range cacheGuilds {
		if guild.ID == e.ID {
			guild.Name = e.Name // For bots, the ready event does not send it's name
			return
		}
	}

	cacheGuilds = append(cacheGuilds, toUserGuild(e.Guild))
}
func guildDelete(session *discordgo.Session, e *discordgo.GuildDelete) {
	index := -1
	for i, guild := range cacheGuilds {
		if guild.ID == e.Guild.ID {
			index = i
		}
	}
	if index >= 0 {
		cacheGuilds = append(cacheGuilds[index:], cacheGuilds[index+1:]...)
	}
}

func toUserGuild(guild *discordgo.Guild) *discordgo.UserGuild {
	return &discordgo.UserGuild{
		ID:   guild.ID,
		Name: guild.Name,
	}
}
