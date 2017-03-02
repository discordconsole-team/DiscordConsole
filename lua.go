package main;

import (
	"github.com/Shopify/go-lua"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

var theSession *discordgo.Session;

func RunLua(session *discordgo.Session, file string, args ...string) error{
	l := lua.NewState();

	l.PushGoFunction(luaExec); l.SetGlobal("exec");
	l.PushGoFunction(luaReplace); l.SetGlobal("replace");
	l.PushGoFunction(luaSleep); l.SetGlobal("sleep");

	l.NewTable();
	for i, val := range args{
		l.PushInteger(i+1);
		l.PushString(val);
		l.SetTable(-3);
	}
	l.SetGlobal("arg");

	lua.OpenLibraries(l);

	theSession = session;

	err := lua.DoFile(l, file);
	return err;
}

func luaExec(l *lua.State) int{
	returnVal := command(theSession, lua.CheckString(l, 1));
	l.PushString(returnVal);
	return 1;
}

func luaReplace(l *lua.State) int{
	replaced := strings.Replace(lua.CheckString(l, 1), lua.CheckString(l, 2), lua.CheckString(l, 3), -1);
	l.PushString(replaced);
	return 1;
}

func luaSleep(l *lua.State) int{
	num := lua.CheckInteger(l, 1);
	time.Sleep(time.Duration(num) * time.Second);
	return 0;
}
