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
	"sync"

	"github.com/bwmarrin/discordgo"
)

var mutexCacheGuilds sync.RWMutex
var cacheGuilds []*discordgo.UserGuild
var cacheChannels []*discordgo.Channel
var cachedChannelType discordgo.ChannelType

var chanReady = make(chan []*discordgo.UserGuild)

func ready(session *discordgo.Session, e *discordgo.Ready) {
	select {
	case _, ok := <-chanReady:
		if !ok {
			return
		}
	default:
	}

	guilds := make([]*discordgo.UserGuild, len(e.Guilds))
	for i, guild := range e.Guilds {
		guilds[i] = toUserGuild(guild)
	}
	if userType == typeUser {
		guilds = sortGuilds(guilds, e.Settings)
	}

	mutexCacheGuilds.Lock()
	cacheGuilds = guilds
	mutexCacheGuilds.Unlock()

	select {
	case chanReady <- guilds:
	default:
	}
	close(chanReady)
}

func guildCreate(session *discordgo.Session, e *discordgo.GuildCreate) {
	mutexCacheGuilds.Lock()
	defer mutexCacheGuilds.Unlock()

	for _, guild := range cacheGuilds {
		if guild.ID == e.ID {
			guild.Name = e.Name
			// For bots, the ready event does not send it's name
			return
		}
	}

	cacheGuilds = append(cacheGuilds, toUserGuild(e.Guild))
}
func guildDelete(session *discordgo.Session, e *discordgo.GuildDelete) {
	mutexCacheGuilds.Lock()
	defer mutexCacheGuilds.Unlock()

	index := -1
	for i, guild := range cacheGuilds {
		if guild.ID == e.Guild.ID {
			index = i
			break
		}
	}
	if index >= 0 {
		// UGH!
		// append(cacheGuilds, cacheGuilds[:index], cacheGuilds[index+1:])
		// would give a pointer leak because Go's garbage collector is ~~stupid~~
		// Garbage* //Mnpn ( ͡° ͜ʖ ͡°)
		copy(cacheGuilds[index:], cacheGuilds[index+1:])
		cacheGuilds[len(cacheGuilds)-1] = nil
		cacheGuilds = cacheGuilds[:len(cacheGuilds)-1]
	}
}

func toUserGuild(guild *discordgo.Guild) *discordgo.UserGuild {
	return &discordgo.UserGuild{
		ID:   guild.ID,
		Name: guild.Name,
	}
}
