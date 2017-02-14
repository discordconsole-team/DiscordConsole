package main;

import (
	"fmt"
	"github.com/legolord208/stdutil"
	"github.com/bwmarrin/discordgo"
	"strings"
	"github.com/legolord208/gtable"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"flag"
	"runtime"
	"sort"
)

const VERSION = "1.7";
const WINDOWS = runtime.GOOS == "windows";
var ID string;

type stringArr []string;

func (arr *stringArr) Set(val string) error{
	*arr = append(*arr, val);
	return nil;
}

func (arr *stringArr) String() string{
	return "[" + strings.Join(*arr, " ") + "]";
}

func main(){
	var token string;
	var email string;
	var pass string;
	var commands stringArr;

	flag.StringVar(&token, "t", "", "Set token. Ignored if -e and/or -p are set.");
	flag.StringVar(&email, "e", "", "Set email.");
	flag.StringVar(&pass, "p", "", "Set password.");
	flag.Var(&commands, "x", "Pre-execute command. Can use flag multiple times.");
	flag.Parse();

	fmt.Println("Discord bot console " + VERSION);
	fmt.Println("Please paste your 'token' here, or leave blank for a user account.");
	fmt.Print("> ");
	if(token == "" && email == "" && pass == ""){
		token = stdutil.MustScanTrim();
	} else{
		if(email != "" || pass != ""){
			token = "";
		}
		fmt.Println(token);
	}

	var session *discordgo.Session;
	var err error;
	if(token == ""){
		fmt.Print("Email: ");
		if(email == ""){
			email = stdutil.MustScanTrim();
		} else {
			fmt.Println(email);
		}
		fmt.Print("Password: ");

		if(pass == ""){
			if(!WINDOWS){
				execute("stty", "-echo");
			}
			pass, err = stdutil.ScanTrim();
			if(!WINDOWS){
				execute("stty", "echo");
				fmt.Println();
			}

			if(err != nil){
				return;
			}
		}

		fmt.Println("Authenticating...");
		session, err = discordgo.New(email, pass);
	} else {
		fmt.Println("Authenticating...");
		session, err = discordgo.New("Bot " + token);
	}

	if(err != nil){
		stdutil.PrintErr("Couldn't authenticate", err);
		return;
	}

	user, err := session.User("@me");
	if(err != nil){
		stdutil.PrintErr("Couldn't query user", err);
		return;
	}

	ID = user.ID;

	session.AddHandler(messageCreate);
	err = session.Open();
	if(err != nil){
		stdutil.PrintErr("Could not open session", err);
	}

	fmt.Println("Logged in with user ID " + ID);
	fmt.Println("Write 'help' for help");
	fmt.Println("Press Ctrl+D or type 'exit' to exit.");
	
	for i := 0; i < 3; i++{
		fmt.Println();
	}

	for _, cmdstr := range commands{
		if(cmdstr == ""){
			continue;
		}
		fmt.Println("> " + cmdstr);

		parts := strings.Fields(cmdstr);
		cmd := parts[0];
		args := parts[1:];
		command(session, cmd, args...);
	}
	for{
		fmt.Print("> ");
		cmdstr, err := stdutil.ScanTrim();
		if(err != nil){
			fmt.Println("exit");
			exit(session);
			return;
		}

		if(cmdstr == ""){
			continue;
		}

		parts := strings.Fields(cmdstr);

		cmd := parts[0];
		args := parts[1:];
		command(session, cmd, args...);
	}
}

func exit(session *discordgo.Session){
	session.Close();
	os.Exit(0);
}

type rolesArr []*discordgo.Role;

func (arr rolesArr) Len() int{
	return len(arr);
}

func (arr rolesArr) Swap(i, j int){
	arr[i], arr[j] = arr[j], arr[i];
}

func (arr rolesArr) Less(i, j int) bool{
	return arr[i].Position > arr[j].Position;
}

type location struct{
	guildID string
	channelID string
}

var loc location;
var lastMsg location;
var lastLoc location;

var cacheGuilds = make(map[string]string, 0);
var cacheChannels = make(map[string]string, 0);

var messages = true;

func command(session *discordgo.Session, cmd string, args... string){
	cmd = strings.ToLower(cmd);
	nargs := len(args);

	if(cmd == "help"){

		fmt.Println("help\tShow help menu");
		fmt.Println("exit\tExit DiscordConsole");
		fmt.Println("exec\tExecute a shell command");
		fmt.Println();
		fmt.Println("guilds\tList guilds/servers this bot is added to.");
		fmt.Println("guild <id>\tSelect a guild to use for further commands.");
		fmt.Println("channels\tList channels in your selected guild.");
		fmt.Println("channel <id>\tSelect a channel to use for further commands.");
		fmt.Println("pchannels\tList private channels a.k.a. 'DMs'.");
		fmt.Println("dm <user id>\tCreate a DM with specific user.");
		fmt.Println();
		fmt.Println("say <stuff>\tSend a message in your selected channel.");
		fmt.Println("file <file>\tUpload file to selected channel.");
		fmt.Println("edit <message id> <stuff>\tEdit a message in your selected channel.");
		fmt.Println("del <message id>\tDelete a message in the selected channel.");
		fmt.Println("delall <since message id>\tBulk delete messages since a specific message");
		fmt.Println("log [output file]\tLog the last few messages in console or to a file.");
		fmt.Println();
		fmt.Println("playing [game]\tSet your playing status.");
		fmt.Println("streaming [url] [game]\tSet your streaming status");
		fmt.Println("typing\tSimulate typing in selected channel...");
		fmt.Println();
		fmt.Println("members\tList (max 100) members in selected guild");
		fmt.Println("invite\tCreate (permanent) instant invite.");
		fmt.Println();
		fmt.Println("roles\tList all roles in selected guild.");
		fmt.Println("roleadd <user id> <role id>\tAdd role to user");
		fmt.Println("roledel <user id> <role id>\tRemove role from user");
		fmt.Println();
		fmt.Println("nick [nickname]\tChange own nicknakme");
		fmt.Println();
		fmt.Println("enablemessages\tEnable intercepting messages");
		fmt.Println("disablemessages\tReverts the above.");
		fmt.Println("reply\tJump to the channel of the last received message.");
		fmt.Println("back\tJump to previous guild and/or channel.");

	} else if(cmd == "exit"){
		exit(session);
	} else if(cmd == "exec" || cmd == "execute"){
		if(nargs < 1){
			stdutil.PrintErr("exec <command>", nil);
			return;
		}

		cmd := strings.Join(args, " ");

		var err error;
		if(WINDOWS){
			err = execute("cmd", "/c", cmd);
		} else {
			err = execute("sh", "-c", cmd);
		}
		if(err != nil){
			stdutil.PrintErr("Could not execute", err);
		}
	} else if(cmd == "guilds"){
		guilds, err := session.UserGuilds();
		if(err != nil){
			stdutil.PrintErr("Could not get guilds", err);
			return;
		}

		cacheGuilds = make(map[string]string, 0);

		table := gtable.NewStringTable();
		table.AddStrings("Name", "ID")

		for _, guild := range guilds{
			table.AddRow();
			table.AddStrings(guild.Name, guild.ID);
			cacheGuilds[strings.ToLower(guild.Name)] = guild.ID;
		}

		printTable(&table);
	} else if(cmd == "guild"){
		if(nargs < 1){
			stdutil.PrintErr("guild <id>", nil);
			return;
		}

		lastLoc = loc;

		var ok bool;
		loc.guildID, ok = cacheGuilds[strings.ToLower(strings.Join(args, " "))];

		if(!ok){
			loc.guildID = args[0];
		}
	} else if(cmd == "channels"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}
		channels, err := session.GuildChannels(loc.guildID);
		if(err != nil){
			stdutil.PrintErr("Could not get channels", nil);
			return;
		}

		cacheChannels = make(map[string]string);

		table := gtable.NewStringTable();
		table.AddStrings("Name", "ID");

		for _, channel := range channels{
			if(channel.Type != "text"){
				continue;
			}
			table.AddRow();
			table.AddStrings(channel.Name, channel.ID);
			cacheChannels[strings.ToLower(channel.Name)] = channel.ID;
		}

		printTable(&table);
	} else if(cmd == "channel"){
		if(nargs < 1){
			stdutil.PrintErr("channel <id>", nil);
			return;
		}

		lastLoc = loc;

		var ok bool;
		loc.channelID, ok = cacheChannels[strings.ToLower(strings.Join(args, " "))];
		if(!ok){
			loc.channelID = args[0];
		}
	} else if(cmd == "say"){
		if(nargs < 1){
			stdutil.PrintErr("say <stuff>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		msg, err := session.ChannelMessageSend(loc.channelID, strings.Join(args, " "));
		if(err != nil){
			stdutil.PrintErr("Could not send", err);
			return;
		}
		fmt.Println("Created message with ID " + msg.ID);
	} else if(cmd == "edit"){
		if(nargs < 2){
			stdutil.PrintErr("edit <message id> <stuff>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		msg, err := session.ChannelMessageEdit(loc.channelID, args[0], strings.Join(args[1:], " "));
		if(err != nil){
			stdutil.PrintErr("Could not edit", err);
			return;
		}
		fmt.Println("Edited " + msg.ID + "!");
	} else if(cmd == "del"){
		if(nargs < 1){
			stdutil.PrintErr("del <message id>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		err := session.ChannelMessageDelete(loc.channelID, args[0]);
		if(err != nil){
			stdutil.PrintErr("Couldn't delete", err);
			return;
		}
	} else if(cmd == "log"){
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		limit := 100;
		if(nargs < 1){
			limit = 10;
		}

		msgs, err := session.ChannelMessages(loc.channelID, limit, "", "");
		if(err != nil){
			stdutil.PrintErr("Could not get messages", err);
			return;
		}
		s := "";

		for i := len(msgs) - 1; i >= 0; i--{
			msg := msgs[i];
			if(msg.Author == nil){
				return;
			}
			s += msg.Author.Username + ": " + msg.Content + "\n";
		}

		if(nargs < 1){
			fmt.Print(s);
			return;
		}

		name := strings.Join(args, " ");
		err = ioutil.WriteFile(name, []byte(s), 0666);
		if(err != nil){
			stdutil.PrintErr("Could not write log file", err);
			return;
		}
		fmt.Println("Wrote chat log to '" + name + "'.")
	} else if(cmd == "playing"){
		err := session.UpdateStatus(0, strings.Join(args, " "));
		if(err != nil){
			stdutil.PrintErr("Couldn't update status", err);
		}
	} else if(cmd == "streaming"){
		var err error;
		if(nargs <= 0){
			err = session.UpdateStreamingStatus(0, "", "");
		} else if(nargs < 2){
			err = session.UpdateStreamingStatus(0, strings.Join(args[1:], " "), "");
		} else {
			err = session.UpdateStreamingStatus(0, strings.Join(args[1:], " "), args[0]);
		}
		if(err != nil){
			stdutil.PrintErr("Couldn't update status", err);
		}
	} else if(cmd == "typing"){
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected.", nil);
			return;
		}
		err := session.ChannelTyping(loc.channelID);
		if(err != nil){
			stdutil.PrintErr("Couldn't start typing", err);
		}
	} else if(cmd == "pchannels"){
		channels, err := session.UserChannels();
		if(err != nil){
			stdutil.PrintErr("Could not get private channels", err);
			return;
		}

		table := gtable.NewStringTable();
		table.AddStrings("ID");

		for _, channel := range channels{
			table.AddRow();
			table.AddStrings(channel.ID);
		}
		printTable(&table);
	} else if(cmd == "dm"){
		if(nargs < 1){
			stdutil.PrintErr("dm <user id>", nil);
			return;
		}
		channel, err := session.UserChannelCreate(args[0]);
		if(err != nil){
			stdutil.PrintErr("Could not create DM.", err);
			return;
		}
		loc.channelID = channel.ID;
		fmt.Println("Selected DM with channel ID " + channel.ID + ".");
	} else if(cmd == "delall"){
		if(nargs < 1){
			stdutil.PrintErr("delall <since message id>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected.", nil);
			return;
		}
		messages, err := session.ChannelMessages(loc.channelID, 100, "", args[0]);
		if(err != nil){
			stdutil.PrintErr("Could not get messages", err);
			return;
		}

		ids := make([]string, len(messages));
		for i, msg := range messages{
			ids[i] = msg.ID;
		}

		err = session.ChannelMessagesBulkDelete(loc.channelID, ids);
		if(err != nil){
			stdutil.PrintErr("Could not delete messages", err);
			return;
		}
		fmt.Println("Deleted " + strconv.Itoa(len(ids)) + " messages!");
	} else if(cmd == "members"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected", nil);
			return;
		}

		members, err := session.GuildMembers(loc.guildID, "", 100);
		if(err != nil){
			stdutil.PrintErr("Could not list members", err);
			return;
		}

		table := gtable.NewStringTable();
		table.AddStrings("Name", "Nick", "ID");

		for _, member := range members{
			table.AddRow();
			table.AddStrings(member.User.Username, member.Nick, member.User.ID);
		}
		printTable(&table);
	} else if(cmd == "invite"){
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected", nil);
			return;
		}
		invite, err := session.ChannelInviteCreate(loc.channelID, discordgo.Invite{});
		if(err != nil){
			stdutil.PrintErr("Invite could not be created", err);
			return;
		}
		fmt.Println("Created invite with code " + invite.Code);
	} else if(cmd == "file"){
		if(nargs < 1){
			stdutil.PrintErr("file <file>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected", nil);
			return;
		}
		name := strings.Join(args, " ");
		file, err := os.OpenFile(name, os.O_RDONLY, 0666);
		if(err != nil){
			stdutil.PrintErr("Couldn't open file", nil);
			return;
		}
		defer file.Close();

		msg, err := session.ChannelFileSend(loc.channelID, name, file);
		if(err != nil){
			stdutil.PrintErr("Could not send file", err);
			return;
		}
		fmt.Println("Sent '" + name + "' with message ID " + msg.ID + ".");
	} else if(cmd == "roles"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected", nil);
			return;
		}

		roles2, err := session.GuildRoles(loc.guildID);
		if(err != nil){
			stdutil.PrintErr("Could not get roles", err);
			return;
		}

		roles := rolesArr(roles2);

		sort.Sort(roles);

		table := gtable.NewStringTable();
		table.AddStrings("Name", "ID", "Permissions");

		for _, role := range roles{
			table.AddRow();
			table.AddStrings(role.Name, role.ID, strconv.Itoa(role.Permissions));
		}

		printTable(&table);
	} else if(cmd == "roleadd" || cmd == "roledel"){
		if(nargs < 2){
			stdutil.PrintErr("roleadd/del <user id> <role id>", nil);
			return;
		}
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected", nil);
			return;
		}

		var err error;
		if(cmd == "roleadd"){
			err = session.GuildMemberRoleAdd(loc.guildID, args[0], args[1]);
		} else {
			err = session.GuildMemberRoleRemove(loc.guildID, args[0], args[1]);
		}

		if(err != nil){
			stdutil.PrintErr("Could not add/remove role", err);
		}
	} else if(cmd == "nick"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected.", nil);
			return;
		}
		err := session.GuildMemberNickname(loc.guildID, "@me/nick", strings.Join(args, " "));
		if(err != nil){
			stdutil.PrintErr("Could not set nickname", err);
		}
	} else if(cmd == "enablemessages"){
		messages = true;
		fmt.Println("Messages will now be intercepted.");
	} else if(cmd == "disablemessages"){
		messages = false;
		fmt.Println("Messages will no longer be intercepted.");
	} else if(cmd == "reply"){
		lastLoc = loc;
		if(lastMsg.guildID != ""){
			loc.guildID = lastMsg.guildID;
		}
		if(lastMsg.channelID != ""){
			loc.channelID = lastMsg.channelID;
		}
	} else if(cmd == "back"){
		loc.guildID = lastLoc.guildID;
		loc.channelID = lastLoc.channelID;
	} else {
		stdutil.PrintErr("Unknown command. Do 'help' for help", nil);
	}
}

func printTable(table *gtable.StringTable){
	table.Each(func(ti *gtable.TableItem){
		ti.Padding(1);
	});
	fmt.Println(table.String());
}

func execute(command string, args... string) error{
	cmd := exec.Command(command, args...);
	cmd.Stdin = os.Stdin;
	cmd.Stdout = os.Stdout;
	cmd.Stderr = os.Stderr;
	return cmd.Run();
}

func messageCreate(session *discordgo.Session, e *discordgo.MessageCreate){
	if(e.Author == nil){}

	if(!messages){
		return;
	}

	s := "\r(";

	channel, err := session.Channel(e.ChannelID);
	if(err != nil){
		fmt.Println("Could not get channel", err);
		return;
	}
	if(channel.IsPrivate){
		s += "Private";
	} else {
		guild, err := session.Guild(channel.GuildID);
		if(err != nil){
			fmt.Println("Could not get guild", err);
			return;
		}
		s += guild.Name + " " + "#" + channel.Name;

		lastMsg.guildID = guild.ID;
	}
	lastMsg.channelID = channel.ID;

	s += ") " + e.Author.Username + ": " + e.Content;
	s += strings.Repeat(" ", 5);
	fmt.Println(s);
	fmt.Print("> ");
}
