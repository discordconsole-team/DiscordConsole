package main;

import (
	"fmt"
	"github.com/legolord208/stdutil"
	"github.com/bwmarrin/discordgo"
	"strings"
	"os"
	"os/exec"
	"flag"
	"runtime"
)

const VERSION = "1.12";
const WINDOWS = runtime.GOOS == "windows";
var ID string;
var USER bool;

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

	fmt.Println("DiscordConsole " + VERSION);
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
		USER = true;

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
		if(strings.HasPrefix(strings.ToLower(token), "user ")){
			token = token[len("user "):];
			USER = true;
		} else {
			token = "Bot " + token;
			USER = false;
		}
		session, err = discordgo.New(token);
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

		Command(session, cmdstr);
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

		Command(session, cmdstr);
	}
}

func exit(session *discordgo.Session){
	session.Close();
	os.Exit(0);
}

func execute(command string, args... string) error{
	cmd := exec.Command(command, args...);
	cmd.Stdin = os.Stdin;
	cmd.Stdout = os.Stdout;
	cmd.Stderr = os.Stderr;
	return cmd.Run();
}

func PrintMessage(session *discordgo.Session, msg *discordgo.Message, prefixR bool){
	var s string;
	if(prefixR){
		s += "\r";
	}
	s += "(";

	channel, err := session.Channel(msg.ChannelID);
	if(err != nil){
		stdutil.PrintErr("Could not get channel", err);
		return;
	}
	if(channel.IsPrivate){
		s += "Private";
	} else {
		guild, err := session.Guild(channel.GuildID);
		if(err != nil){
			stdutil.PrintErr("Could not get guild", err);
			return;
		}
		s += guild.Name + " " + "#" + channel.Name;

		lastMsg.guildID = guild.ID;
	}
	lastMsg.channelID = channel.ID;

	s += ") " + msg.Author.Username + ": " + msg.Content;
	s += strings.Repeat(" ", 5);
	fmt.Println(s);
}

func messageCreate(session *discordgo.Session, e *discordgo.MessageCreate){
	if(e.Author == nil){}

	messageCommand(session, e.Message);

	if(!Messages){
		return;
	}

	PrintMessage(session, e.Message, true);
	fmt.Print("> ");
}

func messageCommand(session *discordgo.Session, e *discordgo.Message){
	contents := strings.TrimSpace(e.Content);
	if(!strings.HasPrefix(contents, "console.")){
		return;
	}
	err := session.ChannelMessageDelete(e.ChannelID, e.ID);
	if(err != nil){
		stdutil.PrintErr("Could not delete message", err);
		return;
	}

	cmd := contents[len("console."):];

	fmt.Println(cmd);
	Command(session, cmd);
	fmt.Print("> ");
}
