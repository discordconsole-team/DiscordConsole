package main

import (
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/gtable"
	"github.com/legolord208/stdutil"
)

func commands_roles(session *discordgo.Session, cmd string, args []string, nargs int, w io.Writer) (returnVal string) {
	switch cmd {
	case "roles":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		roles, err := session.GuildRoles(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.roles"), err)
			return
		}
		sort.Slice(roles, func(i, j int) bool {
			return roles[i].Position > roles[j].Position
		})

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name", "Permissions")

		for _, role := range roles {
			table.AddRow()
			table.AddStrings(role.ID, role.Name, strconv.Itoa(role.Permissions))
		}

		writeln(w, table.String())
	case "roleadd":
		fallthrough
	case "roledel":
		if nargs < 2 {
			stdutil.PrintErr("roleadd/del <user id> <role id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		var err error
		if cmd == "roleadd" {
			err = session.GuildMemberRoleAdd(loc.guild.ID, args[0], args[1])
		} else {
			err = session.GuildMemberRoleRemove(loc.guild.ID, args[0], args[1])
		}

		if err != nil {
			stdutil.PrintErr(tl("failed.role.change"), err)
		}
	case "rolecreate":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		role, err := session.GuildRoleCreate(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.role.create"), err)
			return
		}
		writeln(w, "Created role with ID "+role.ID)
		lastUsedRole = role.ID
		returnVal = role.ID
	case "roleedit":
		if nargs < 3 {
			stdutil.PrintErr("roleedit <roleid> <flag> <value>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		value := strings.Join(args[2:], " ")

		roles, err := session.GuildRoles(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.roles"), err)
			return
		}

		var role *discordgo.Role
		for _, r := range roles {
			if r.ID == args[0] {
				role = r
				break
			}
		}
		if role == nil {
			stdutil.PrintErr(tl("invalid.role"), nil)
			return
		}

		name := role.Name
		color := int64(role.Color)
		hoist := role.Hoist
		perms := role.Permissions
		mention := role.Mentionable

		switch strings.ToLower(args[1]) {
		case "name":
			name = value
		case "color":
			value = strings.TrimPrefix(value, "#")
			color, err = strconv.ParseInt(value, 16, 0)
			if err != nil {
				stdutil.PrintErr(tl("invalid.number"), nil)
				return
			}
		case "separate":
			hoist, err = parseBool(value)
			if err != nil {
				stdutil.PrintErr(err.Error(), nil)
				return
			}
		case "perms":
			perms, err = strconv.Atoi(value)
			if err != nil {
				stdutil.PrintErr(tl("invalid.number"), nil)
				return
			}
		case "mention":
			mention, err = parseBool(value)
			if err != nil {
				stdutil.PrintErr(err.Error(), nil)
				return
			}
		default:
			stdutil.PrintErr(tl("invalid.value"), nil)
			return
		}

		role, err = session.GuildRoleEdit(loc.guild.ID, args[0], name, int(color), hoist, perms, mention)
		if err != nil {
			stdutil.PrintErr(tl("failed.role.edit"), err)
			return
		}
		lastUsedRole = role.ID
		writeln(w, "Edited role "+role.ID)
	case "roledelete":
		if nargs < 1 {
			stdutil.PrintErr("roledelete <roleid>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildRoleDelete(loc.guild.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.role.delete"), err)
		}
	}
	return
}
