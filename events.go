package main

import "github.com/bwmarrin/discordgo"

var cacheGuilds []*discordgo.UserGuild
var cacheChannels []*discordgo.Channel
var cachedChannelType string

var chanReady = make(chan []*discordgo.UserGuild)

func ready(session *discordgo.Session, e *discordgo.Ready) {
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
