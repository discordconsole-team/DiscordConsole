package main;

import (
	"github.com/Shopify/go-lua"
	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/stdutil"
	"strings"
)

var theSession *discordgo.Session;

func RunLua(session *discordgo.Session, file string) error{
	l := lua.NewState();

	l.PushGoFunction(send); l.SetGlobal("exec");
	l.PushGoFunction(read); l.SetGlobal("read");
	l.PushGoFunction(replace); l.SetGlobal("replace");

	lua.OpenLibraries(l);

	theSession = session;

	err := lua.DoFile(l, file);
	return err;
}

func send(l *lua.State) int{
	returnVal := Command(theSession, lua.CheckString(l, 1));
	l.PushString(returnVal);
	return 1;
}

func read(l *lua.State) int{
	in, _ := stdutil.ScanTrim();
	l.PushString(in);
	return 1;
}

func replace(l *lua.State) int{
	replaced := strings.Replace(lua.CheckString(l, 1), lua.CheckString(l, 2), lua.CheckString(l, 3), -1);
	l.PushString(replaced);
	return 1;
}
