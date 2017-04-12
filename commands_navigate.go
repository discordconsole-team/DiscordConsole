package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/gtable"
	"github.com/legolord208/stdutil"
)

func commands_navigate(session *discordgo.Session, cmd string, args []string, nargs int) (returnVal string) {
	switch cmd {
	case "guilds":
		guilds, err := session.UserGuilds(100, "", "")
		if err != nil {
			stdutil.PrintErr(tl("failed.guild"), err)
			return
		}

		cacheGuilds = make(map[string]string)

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name")

		for _, guild := range guilds {
			table.AddRow()
			table.AddStrings(guild.ID, guild.Name)
			cacheGuilds[strings.ToLower(guild.Name)] = guild.ID
		}

		printTable(table)
	case "guild":
		if nargs < 1 {
			stdutil.PrintErr("guild <id>", nil)
			return
		}

		guildID, ok := cacheGuilds[strings.ToLower(strings.Join(args, " "))]
		if !ok {
			guildID = args[0]
		}

		guild, err := session.Guild(guildID)
		if err != nil {
			stdutil.PrintErr(tl("failed.guild"), err)
			return
		}

		channel, err := session.Channel(guildID)
		if err != nil {
			stdutil.PrintErr(tl("failed.channel"), err)
			return
		}
		loc.push(guild, channel)
	case "channels":
		channels(session, "text")
	case "channel":
		if nargs < 1 {
			stdutil.PrintErr("channel <id>", nil)
			return
		}

		channelID, ok := cacheChannels[strings.ToLower(strings.Join(args, " "))]
		if !ok {
			channelID = args[0]
		}

		channel, err := session.Channel(channelID)
		if err != nil {
			stdutil.PrintErr(tl("failed.channel"), err)
			return
		}
		if channel.IsPrivate {
			loc.push(nil, channel)
		} else {
			if loc.guild == nil || channel.GuildID != loc.guild.ID {
				guild, err := session.Guild(channel.GuildID)

				if err != nil {
					stdutil.PrintErr(tl("failed.guild"), err)
					return
				}

				loc.push(guild, channel)
			} else {
				loc.push(loc.guild, channel)
			}
		}
	case "pchannels":
		channels, err := session.UserChannels()
		if err != nil {
			stdutil.PrintErr(tl("failed.channel"), err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Recipient")

		for _, channel := range channels {
			table.AddRow()
			recipient := ""
			if channel.Recipient != nil {
				recipient = channel.Recipient.Username
			}
			table.AddStrings(channel.ID, recipient)
		}
		printTable(table)
	case "dm":
		if nargs < 1 {
			stdutil.PrintErr("dm <user id>", nil)
			return
		}
		channel, err := session.UserChannelCreate(args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.channel.create"), err)
			return
		}
		loc.push(nil, channel)

		fmt.Println(tl("channel.select") + " " + channel.ID)
	case "bookmarks":
		for key, _ := range bookmarks {
			fmt.Println(key)
		}
	case "bookmark":
		if nargs < 1 {
			stdutil.PrintErr("bookmark <name>", nil)
			return
		}

		key := strings.Join(args, " ")
		if strings.HasPrefix(key, "-") {
			key = key[1:]
			delete(bookmarks, key)
		} else {
			bookmarks[key] = loc.channel.ID
		}
		err := saveBookmarks()
		if err != nil {
			stdutil.PrintErr(tl("failed.file.save"), err)
		}
	case "go":
		if nargs < 1 {
			stdutil.PrintErr("go <bookmark>", nil)
			return
		}
		bookmark, ok := bookmarks[args[0]]
		if !ok {
			stdutil.PrintErr(tl("invalid.bookmark"), nil)
			return
		}

		var guild *discordgo.Guild
		var channel *discordgo.Channel
		var err error

		if bookmark != "" {
			channel, err = session.Channel(bookmark)
			if err != nil {
				stdutil.PrintErr(tl("failed.channel"), err)
				return
			}
		}

		if channel != nil && !channel.IsPrivate {
			guild, err = session.Guild(channel.GuildID)
			if err != nil {
				stdutil.PrintErr(tl("failed.guild"), err)
				return
			}
		}

		loc.push(guild, channel)
	}
	return
}

func channels(session *discordgo.Session, kind string) {
	if loc.guild == nil {
		stdutil.PrintErr(tl("invalid.guild"), nil)
		return
	}
	channels, err := session.GuildChannels(loc.guild.ID)
	if err != nil {
		stdutil.PrintErr(tl("failed.channel"), nil)
		return
	}

	cacheChannels = make(map[string]string)

	sort.Slice(channels, func(i int, j int) bool {
		return channels[i].Position < channels[j].Position
	})

	table := gtable.NewStringTable()
	table.AddStrings("ID", "Name")

	for _, channel := range channels {
		if channel.Type != kind {
			continue
		}
		table.AddRow()
		table.AddStrings(channel.ID, channel.Name)
		cacheChannels[strings.ToLower(channel.Name)] = channel.ID
	}

	printTable(table)
}
