package main

import (
	"io"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/gtable"
	"github.com/legolord208/stdutil"
)

func commands_navigate(session *discordgo.Session, cmd string, args []string, nargs int, w io.Writer) (returnVal string) {
	switch cmd {
	case "guilds":
		var guilds []*discordgo.UserGuild
		if cacheGuilds != nil {
			guilds = cacheGuilds
		} else {
			var err error
			guilds, err = session.UserGuilds(100, "", "")
			if err != nil {
				stdutil.PrintErr(tl("failed.guild"), err)
				return
			}

			if UserType == TypeUser {
				settings, err := session.UserSettings()
				if err != nil {
					stdutil.PrintErr(tl("failed.settings"), err)
				} else {
					guilds2 := guilds

					guilds = make([]*discordgo.UserGuild, len(settings.GuildPositions))
					for i, g := range settings.GuildPositions {
						for _, g2 := range guilds2 {
							if g == g2.ID {
								guilds[i] = g2
							}
						}
					}

					// Should never happen, if the two endpoints are in sync.
					// But we want to avoid any crash at all costs.
					for i, g := range guilds {
						if g == nil {
							guilds[i] = &discordgo.UserGuild{Name: "Error"}
						}
					}
				}
			}

			cacheGuilds = guilds
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name")

		for _, guild := range guilds {
			table.AddRow()
			table.AddStrings(guild.ID, guild.Name)
		}

		writeln(w, table.String())
	case "guild":
		if nargs < 1 {
			stdutil.PrintErr("guild <id>", nil)
			return
		}

		guildID := strings.Join(args, " ")
		for _, g := range cacheGuilds {
			if strings.EqualFold(guildID, g.Name) {
				guildID = g.ID
				break
			}
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
		channels(session, "text", w)
	case "channel":
		if nargs < 1 {
			stdutil.PrintErr("channel <id>", nil)
			return
		}

		arg := strings.Join(args, " ")

		var channel *discordgo.Channel
		for _, c := range cacheChannels {
			if strings.EqualFold(arg, c.Name) {
				channel = c
			}
		}
		if channel == nil {
			var err error
			channel, err = session.Channel(arg)
			if err != nil {
				stdutil.PrintErr(tl("failed.channel"), err)
				return
			}
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
		writeln(w, table.String())
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

		writeln(w, tl("channel.select")+" "+channel.ID)
	case "bookmarks":
		for key, _ := range bookmarks {
			writeln(w, key)
		}
	case "bookmark":
		if nargs < 1 {
			stdutil.PrintErr("bookmark <name>", nil)
			return
		}

		name := strings.ToLower(strings.Join(args, " "))
		if strings.HasPrefix(name, "-") {
			name = name[1:]
			delete(bookmarks, name)
			delete(bookmarksCache, name)
		} else {
			bookmarks[name] = loc.channel.ID
			bookmarksCache[name] = loc
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
		name := strings.ToLower(strings.Join(args, " "))
		if cache, ok := bookmarksCache[name]; ok {
			loc.push(cache.guild, cache.channel)
			return
		}

		bookmark, ok := bookmarks[name]
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

		bookmarksCache[name] = location{
			guild:   guild,
			channel: channel,
		}

		loc.push(guild, channel)
	}
	return
}

func channels(session *discordgo.Session, kind string, w io.Writer) {
	var channels []*discordgo.Channel
	if cacheChannels != nil && cachedChannelType == kind {
		channels = cacheChannels
	} else {
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}
		channels2, err := session.GuildChannels(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.channel"), nil)
			return
		}

		cacheChannels = channels
		cachedChannelType = kind

		channels = make([]*discordgo.Channel, 0)
		for _, c := range channels2 {
			if c.Type != kind {
				continue
			}
			channels = append(channels, c)
		}

		sort.Slice(channels, func(i int, j int) bool {
			return channels[i].Position < channels[j].Position
		})

		cacheChannels = channels
		cachedChannelType = kind
	}

	table := gtable.NewStringTable()
	table.AddStrings("ID", "Name")

	for _, channel := range channels {
		table.AddRow()
		table.AddStrings(channel.ID, channel.Name)
	}

	writeln(w, table.String())
}
