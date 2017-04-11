package main

import (
	"github.com/Shopify/go-lua"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/legolord208/stdutil"
	"strconv"
	"strings"
	"time"
)

type luaEventData struct {
	state    *lua.State
	function string
}

var luaSessionCopy *discordgo.Session
var luaMessageEvents = make(map[string]*luaEventData)

func RunLua(session *discordgo.Session, file string, args ...string) error {
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

	luaSessionCopy = session

	err := lua.DoFile(l, file)
	return err
}

func luaExec(l *lua.State) int {
	ColorAutomated.Set()
	returnVal := command(luaSessionCopy, lua.CheckString(l, 1))
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
		stdutil.PrintErr(lang["failed.timestamp"], err)
	}

	defer func() {
		r := recover()
		if r != nil {
			stdutil.PrintErr(lang["failed.lua.event"], nil)
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

	luaSessionCopy = session
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
