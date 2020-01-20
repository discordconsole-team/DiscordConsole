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
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/go-lua"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/jD91mZM2/stdutil"
)

type luaEventData struct {
	state    *lua.State
	function string
}

var luaMessageEvents = make(map[string]*luaEventData)

func runLua(session *discordgo.Session, file string, args ...string) error {
	l := lua.NewState()

	l.Register("exec", luaExec)
	l.Register("replace", luaReplace)
	l.Register("sleep", luaSleep)
	l.Register("registerEvent", luaRegister)

	l.NewTable()
	for i, val := range args {
		l.PushInteger(i + 1)
		l.PushString(val)
		l.SetTable(-3)
	}
	l.SetGlobal("arg")

	lua.OpenLibraries(l)

	err := lua.DoFile(l, file)
	return err
}

func luaExec(l *lua.State) int {
	colorAutomated.Set()
	returnVal := command(session, commandSource{}, lua.CheckString(l, 1), color.Output)
	color.Unset()

	l.PushString(returnVal)
	return 1
}
func luaReplace(l *lua.State) int {
	replaced := strings.Replace(lua.CheckString(l, 1), lua.CheckString(l, 2), lua.CheckString(l, 3), -1)
	l.PushString(replaced)
	return 1
}
func luaSleep(l *lua.State) int {
	num := lua.CheckInteger(l, 1)
	time.Sleep(time.Duration(num) * time.Second)
	return 0
}
func luaRegister(l *lua.State) int {
	id := lua.CheckString(l, 1)
	name := lua.CheckString(l, 2)
	luaMessageEvents[id] = &luaEventData{
		state:    l,
		function: name,
	}
	return 0
}

func luaMessageEvent(session *discordgo.Session, e *discordgo.Message) {
	timestamp, err := timestamp(e)
	if err != nil {
		stdutil.PrintErr(tl("failed.timestamp"), err)
	}

	defer func() {
		r := recover()
		if r != nil {
			stdutil.PrintErr(tl("failed.lua.event"), nil)
		}
	}()

	params := map[string]string{
		"ID":        e.ID,
		"Content":   e.Content,
		"ChannelID": e.ChannelID,
		"Timestamp": timestamp,

		"AuthorID":     e.Author.ID,
		"AuthorBot":    strconv.FormatBool(e.Author.Bot),
		"AuthorAvatar": e.Author.Avatar,
		"AuthorName":   e.Author.Username,
	}

	for _, event := range luaMessageEvents {
		l := event.state
		l.Global(event.function)

		l.NewTable()
		for key, val := range params {
			l.PushString(key)
			l.PushString(val)
			l.SetTable(-3)
		}

		l.Call(1, 0)
	}
}
