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
	"os/signal"
	"syscall"
)

const VERSION = "1.16";
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

	if(token == "" && email == "" && pass == ""){
		foundtoken, err := findToken();
		if(err == nil){
			for{
				fmt.Print("You are logged into Discord. Use that login? (y/n): ");
				response := stdutil.MustScanTrim();
				if(strings.EqualFold(response, "y")){
					foundtoken = strings.TrimPrefix(foundtoken, "\"");
					foundtoken = strings.TrimSuffix(foundtoken, "\"");
					token = "user " + foundtoken;
				} else if(!strings.EqualFold(response, "n")){
					fmt.Println("Please type either 'y' or 'n'.");
					continue;
				}
				break;
			}
		}
	}

	fmt.Println("Please paste your 'token' here, or leave blank for a username/password prompt.");
	fmt.Print("> ");
	if(token == "" && email == "" && pass == ""){
		token = stdutil.MustScanTrim();
	} else{
		if(email != "" || pass != ""){
			token = "";
		}
		fmt.Println("[CENSORED]");
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

	go func(){
		interrupt := make(chan os.Signal, 1);
		signal.Notify(interrupt, os.Interrupt);

		term := make(chan os.Signal, 1);
		signal.Notify(term, syscall.SIGTERM);

		for{
			select{
				case <-interrupt:
					fmt.Println();
					fmt.Println("Press Ctrl+D or type 'exit' to exit.");
					printPointer(session);
				case <-term:
					exit(session);
					return;
			}
		}
	}();
	go func(){
	}();

	for _, cmdstr := range commands{
		if(cmdstr == ""){
			continue;
		}
		printPointer(session);
		fmt.Println(cmdstr);

		command(session, cmdstr);
	}
	for{
		printPointer(session);
		cmdstr, err := stdutil.ScanTrim();
		if(err != nil){
			fmt.Println("exit");
			exit(session);
			return;
		}

		if(cmdstr == ""){
			continue;
		}

		command(session, cmdstr);
	}
}

func exit(session *discordgo.Session){
	playing = "";
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

func PrintMessage(session *discordgo.Session, msg *discordgo.Message, prefixR bool, channel *discordgo.Channel){
	var s string;
	if(prefixR){
		s += "\r";
	}
	s += "(";

	var err error;
	if(channel == nil){
		channel, err = session.Channel(msg.ChannelID);
		if(err != nil){
			stdutil.PrintErr("Could not get channel", err);
			return;
		}
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
	}

	s += ") " + msg.Author.Username + ": " + msg.Content;
	s += strings.Repeat(" ", 5);
	fmt.Println(s);
}

func messageCreate(session *discordgo.Session, e *discordgo.MessageCreate){
	if(e.Author == nil){}

	channel, err := session.Channel(e.ChannelID);
	if(err != nil){
		stdutil.PrintErr("Could not get channel", err);
		return;
	}

	if(messageCommand(session, e.Message, channel)){
		return;
	}

	lastMsg.channelID = e.ChannelID;
	lastMsg.guildID = channel.GuildID;

	if(messages){
		PrintMessage(session, e.Message, true, channel);
		printPointer(session);
	}
}

func messageCommand(session *discordgo.Session, e *discordgo.Message, channel *discordgo.Channel) bool{
	if(e.Author.ID != ID){
		return false;
	} else if(!intercept){
		return false;
	}

	contents := strings.TrimSpace(e.Content);
	if(!strings.HasPrefix(contents, "console.")){
		return false;
	}

	err := session.ChannelMessageDelete(e.ChannelID, e.ID);
	if(err != nil){
		stdutil.PrintErr("Could not delete message", err);
	}

	lastLoc = loc;
	loc.channelID = e.ChannelID;
	loc.guildID = channel.GuildID;
	pointerCache = "";

	cmd := contents[len("console."):];

	fmt.Println(cmd);
	command(session, cmd);
	printPointer(session);
	return true;
}

const ERROR_POINTER = "Error> ";
var pointerCache string;

func clearPointerCache(){
	pointerCache = "";
}
func errorPointer(session *discordgo.Session){
	pointerCache = ERROR_POINTER;
	fmt.Print(ERROR_POINTER);
}
func printPointer(session *discordgo.Session){
	if(pointerCache != ""){
		fmt.Print(pointerCache);
		return;
	}

	if(loc.channelID == ""){
		fmt.Print("> ");
		return;
	}

	s := "";

	channel, err := session.Channel(loc.channelID);
	if(err != nil){
		stdutil.PrintErr("Could not get channel", err);
		errorPointer(session);
		return;
	}

	if(channel.IsPrivate){
		s += "Private";
	} else {
		guild, err := session.Guild(loc.guildID);
		if(err != nil){
			stdutil.PrintErr("Could not get guild", err);
			errorPointer(session);
			return;
		}
		s += guild.Name + " (#" + channel.Name + ")";
	}

	s += "> ";
	fmt.Print(s);
	pointerCache = s;
}
