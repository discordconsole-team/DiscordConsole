package main;

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/legolord208/stdutil"
	"github.com/bwmarrin/discordgo"
	"strings"
	"os"
	"os/exec"
	"flag"
	"runtime"
	"os/signal"
	"syscall"
	"io"
	"github.com/fatih/color"
	"io/ioutil"
)

const VERSION = "1.20.1";
const AUTORUN_FILE = ".autorun";
var ID string;
var USER bool;

const WINDOWS = runtime.GOOS == "windows";
const MAC = runtime.GOOS == "darwin";

var READLINE *readline.Instance;
var COLOR_DEFAULT = color.New(color.Bold);
var COLOR_AUTOMATED = color.New(color.Italic);

type stringArr []string;

func (arr *stringArr) Set(val string) error{
	*arr = append(*arr, val);
	return nil;
}

func (arr *stringArr) String() string{
	return "[" + strings.Join(*arr, " ") + "]";
}

var nopointer bool;
var noguildmatch bool;
func main(){
	var token string;
	var email string;
	var pass string;
	var commands stringArr;

	var noupdate bool;
	var noautorun bool;

	flag.StringVar(&token, "t", "", "Set token. Ignored if -e and/or -p are set.");
	flag.StringVar(&email, "e", "", "Set email.");
	flag.StringVar(&pass, "p", "", "Set password.");
	flag.Var(&commands, "x", "Pre-execute command. Can use flag multiple times.");

	flag.BoolVar(&noupdate, "noupdate", false, "Disable update checking.");
	flag.BoolVar(&noautorun, "noautorun", false, "Disable running commands in " + AUTORUN_FILE + " file.");

	flag.BoolVar(&nopointer, "nopointer", false, "Disable pointer (only reason would be speed, I guess).");
	flag.BoolVar(&noguildmatch, "noguildmatch", false, "Disable guild and channel matching (only reason would be speed, I guess).");
	flag.Parse();

	stdutil.ErrOutput = os.Stdout;
	stdutil.EventPrePrintError = append(stdutil.EventPrePrintError, func(text string) bool{
		color.Unset();
		color.Set(color.FgRed, color.Bold);
		return false;
	});
	stdutil.EventPostPrintError = append(stdutil.EventPostPrintError, func(text string){
		color.Unset();
	});

	fmt.Println("DiscordConsole " + VERSION);

	if(!noupdate){
		fmt.Print("Checking for updates... ");
		update, err := checkUpdate();
		if(err != nil){
			stdutil.PrintErr("Error checking for updates", err);
			return;
		}
		if(update.UpdateAvailable){
			fmt.Println();
			color.Cyan("Update available: Version " + update.Version + ".");
			color.Cyan("Download from " + update.Url + ".");
		} else {
			fmt.Println("No updates found.");
		}
	}

	fmt.Println("Reading bookmarks...");
	err := loadBookmarks();
	if(err != nil){
		stdutil.PrintErr("Could not read bookmarks", err);
	}

	var ar_lines []string;
	if(!noautorun){
		ar, err := ioutil.ReadFile(AUTORUN_FILE);
		if(err != nil && os.IsExist(err)){
			stdutil.PrintErr("Could not read " + AUTORUN_FILE, err);
		} else if(err == nil){
			ar_lines = strings.Split(string(ar), "\n");

			if(len(ar_lines) > 0){
				first_line := ar_lines[0];
				if(strings.HasPrefix(first_line, ":")){
					token = first_line[1:];
					ar_lines = ar_lines[1:];
				}
			}
		}
	}

	READLINE, err = readline.New(EMPTY_POINTER);
	if(err != nil){
		stdutil.PrintErr("Could not start readline library", err);
		return;
	}

	if(token == "" && email == "" && pass == ""){
		foundtoken, err := findToken();
		if(err == nil){
			for{
				color.Set(color.FgYellow);
				fmt.Print("You are logged into Discord. Use that login? (y/n): ");
				response := stdutil.MustScanTrim();
				color.Unset();

				if(strings.EqualFold(response, "y")){
					foundtoken = strings.TrimPrefix(foundtoken, "\"");
					foundtoken = strings.TrimSuffix(foundtoken, "\"");
					token = "user " + foundtoken;
				} else if(!strings.EqualFold(response, "n")){
					stdutil.PrintErr("Please type either 'y' or 'n'.", nil);
					continue;
				}
				break;
			}
		}
	}

	fmt.Println("Please paste your 'token' here, or leave blank for a username/password prompt.");
	fmt.Print("> ");
	if(token == "" && email == "" && pass == ""){
		token, err = READLINE.Readline();
		if(err != nil){
			if(err != io.EOF){
				stdutil.PrintErr("Could not read line", err);
			}
			return;
		}
	} else{
		if(email != "" || pass != ""){
			token = "";
		}
		fmt.Println("[CENSORED]");
	}

	var session *discordgo.Session;
	if(token == ""){
		USER = true;

		READLINE.SetPrompt("Email: ");
		if(email == ""){
			email, err = READLINE.Readline();
		} else {
			fmt.Println(email);
		}

		if(pass == ""){
			pass2, err := READLINE.ReadPassword("Password: ");
			fmt.Println();

			if(err != nil){
				if(err != io.EOF){
					stdutil.PrintErr("Could not read password", err);
				}
				return;
			}
			pass = string(pass2);
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
		c := make(chan os.Signal, 1);
		signal.Notify(c, os.Interrupt, syscall.SIGTERM);

		for _ = range c{
			exit(session);
			return;
		}
	}();

	COLOR_AUTOMATED.Set();

	if(ar_lines != nil){
		for _, cmd := range ar_lines{
			cmd = strings.TrimSpace(cmd);
			if(cmd == ""){
				continue;
			}
			printPointer(session);
			fmt.Println(cmd);

			command(session, cmd);
		}
	}
	for _, cmd := range commands{
		cmd = strings.TrimSpace(cmd);
		if(cmd == ""){
			continue;
		}
		printPointer(session);
		fmt.Println(cmd);

		command(session, cmd);
	}

	color.Unset();
	setCompleter(READLINE);

	for{
		COLOR_DEFAULT.Set();

		READLINE.SetPrompt(pointer(session));
		cmd, err := READLINE.Readline();

		color.Unset();

		if(err != nil){
			if(err != io.EOF && err != readline.ErrInterrupt){
				stdutil.PrintErr("Could not read line", err);
			} else {
				fmt.Println("exit");
			}
			exit(session);
			return;
		}

		cmd = strings.TrimSpace(cmd);
		if(cmd == ""){
			continue;
		}

		command(session, cmd);
	}
}

func exit(session *discordgo.Session){
	color.Unset();
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

func printMessage(session *discordgo.Session, msg *discordgo.Message, prefixR bool, guild *discordgo.Guild, channel *discordgo.Channel){
	var s string;
	if(prefixR){
		s += "\r";
	}
	s += "(";

	if(channel.IsPrivate){
		s += "Private";
	} else {
		s += guild.Name + " " + "#" + channel.Name;
	}

	s += ") " + msg.Author.Username + ": " + msg.Content;
	s += strings.Repeat(" ", 5);

	color.Unset();
	color.Yellow(s);
	COLOR_DEFAULT.Set();
}

func messageCreate(session *discordgo.Session, e *discordgo.MessageCreate){
	if(e.Author == nil){}

	channel, err := session.Channel(e.ChannelID);
	if(err != nil){
		stdutil.PrintErr("Could not get channel", err);
		return;
	}

	var guild *discordgo.Guild;
	if(!channel.IsPrivate){
		guild, err = session.Guild(channel.GuildID);
		if(err != nil){
			stdutil.PrintErr("Could not get guild", err);
			return;
		}
	}

	if(messageCommand(session, e.Message, guild, channel)){
		return;
	}

	lastMsg = location{
		guild: guild,
		channel: channel,
	};

	hasOutput := false;

	if(messages){
		printMessage(session, e.Message, true, guild, channel);
		hasOutput = true;
	}

	if(len(luaMessageEvents) > 0){
		hasOutput = true;

		color.Unset();
		COLOR_AUTOMATED.Set();

		fmt.Print("\r" + strings.Repeat(" ", 20) + "\r");
		luaMessageEvent(session, e.Message);

		color.Unset();
		COLOR_DEFAULT.Set();
	}
	if(hasOutput){
		printPointer(session);
	}
}

func messageCommand(session *discordgo.Session, e *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel) bool{
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
	loc = location{
		guild: guild,
		channel: channel,
	};
	pointerCache = "";

	cmd := contents[len("console."):];

	color.Unset();
	COLOR_AUTOMATED.Set();

	fmt.Println(cmd);
	command(session, cmd);

	color.Unset();
	COLOR_DEFAULT.Set();

	printPointer(session);
	return true;
}

const EMPTY_POINTER = "> ";
var pointerCache string;

func printPointer(session *discordgo.Session){
	fmt.Print(pointer(session));
}
func pointer(session *discordgo.Session) string{
	if(pointerCache != ""){
		return pointerCache;
	}

	if(nopointer){
		return EMPTY_POINTER;
	}
	if(loc.channel == nil){
		return EMPTY_POINTER;
	}

	s := "";

	if(loc.channel.IsPrivate){
		s += "Private (" + loc.channel.Recipient.Username + ")";
	} else {
		s += loc.guild.Name + " (#" + loc.channel.Name + ")";
	}

	s += EMPTY_POINTER;
	pointerCache = s;
	return s;
}
