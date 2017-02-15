package main;

import (
	"github.com/Shopify/go-lua"
	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/stdutil"
)

var theSession *discordgo.Session;

func RunLua(session *discordgo.Session, file string) error{
	l := lua.NewState();

	l.PushGoFunction(send);
	l.SetGlobal("exec");

	l.PushGoFunction(read);
	l.SetGlobal("read");

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
