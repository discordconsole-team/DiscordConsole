package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/legolord208/stdutil"
)

const VERSION = "1.22.2"

var DEV_VERSION = strings.Contains(VERSION, "dev")

const AUTORUN_FILE = ".autorun"

var ID string
var USER bool

var READLINE *readline.Instance
var COLOR_DEFAULT = color.New(color.Bold)
var COLOR_AUTOMATED = color.New(color.Italic)
var COLOR_ERROR = color.New(color.FgRed, color.Bold)

const MSG_LIMIT = 2000

type stringArr []string

func (arr *stringArr) Set(val string) error {
	*arr = append(*arr, val)
	return nil
}

func (arr *stringArr) String() string {
	return "[" + strings.Join(*arr, " ") + "]"
}

func main() {
	var token string
	var email string
	var pass string
	var commands stringArr

	var noupdate bool
	var noautorun bool

	flag.StringVar(&token, "t", "", "Set token. Ignored if -e and/or -p are set.")
	flag.StringVar(&email, "e", "", "Set email.")
	flag.StringVar(&pass, "p", "", "Set password.")
	flag.Var(&commands, "x", "Pre-execute command. Can use flag multiple times.")

	flag.BoolVar(&noupdate, "noupdate", false, "Disable update checking.")
	flag.BoolVar(&noautorun, "noautorun", false, "Disable running commands in "+AUTORUN_FILE+" file.")
	flag.Parse()

	doHook()
	fmt.Println("DiscordConsole " + VERSION)

	if !noupdate {
		fmt.Print("Checking for updates... ")
		update, err := checkUpdate()
		if err != nil {
			stdutil.PrintErr("Error checking for updates", err)
			return
		}
		if update.UpdateAvailable {
			fmt.Println()
			if DEV_VERSION {
				color.Cyan("Latest stable release: " + update.Version + ".")
			} else {
				color.Cyan("Update available: Version " + update.Version + ".")
			}
			color.Cyan("Download from " + update.Url + ".")
		} else {
			fmt.Println("No updates found.")
		}
	}

	fmt.Println("Reading bookmarks...")
	err := loadBookmarks()
	if err != nil {
		stdutil.PrintErr("Could not read bookmarks", err)
	}

	var ar_lines []string
	if !noautorun {
		ar, err := ioutil.ReadFile(AUTORUN_FILE)
		if err != nil && os.IsExist(err) {
			stdutil.PrintErr("Could not read "+AUTORUN_FILE, err)
		} else if err == nil {
			ar_lines = strings.Split(string(ar), "\n")

			if len(ar_lines) > 0 {
				first_line := ar_lines[0]
				if strings.HasPrefix(first_line, ":") {
					token = first_line[1:]
					ar_lines = ar_lines[1:]
				}
			}
		}
	}

	READLINE, err = readline.New(EMPTY_POINTER)
	if err != nil {
		stdutil.PrintErr("Could not start readline library", err)
		return
	}

	if token == "" && email == "" && pass == "" {
		foundtoken, err := findToken()
		if err == nil {
			for {
				color.Set(color.FgYellow)
				fmt.Print("You are logged into Discord. Use that login? (y/n): ")
				response := stdutil.MustScanTrim()
				color.Unset()

				if strings.EqualFold(response, "y") {
					foundtoken = strings.TrimPrefix(foundtoken, "\"")
					foundtoken = strings.TrimSuffix(foundtoken, "\"")
					token = "user " + foundtoken
				} else if !strings.EqualFold(response, "n") {
					stdutil.PrintErr("Please type either 'y' or 'n'.", nil)
					continue
				}
				break
			}
		}
	}

	fmt.Println("Please paste your 'token' here, or leave blank for a username/password prompt.")
	fmt.Print("> ")
	if token == "" && email == "" && pass == "" {
		token, err = READLINE.Readline()
		if err != nil {
			if err != io.EOF && err != readline.ErrInterrupt {
				stdutil.PrintErr("Could not read line", err)
			}
			return
		}
	} else {
		if email != "" || pass != "" {
			token = ""
		}
		fmt.Println("[CENSORED]")
	}

	var session *discordgo.Session
	if token == "" {
		USER = true

		READLINE.SetPrompt("Email: ")
		if email == "" {
			email, err = READLINE.Readline()
		} else {
			fmt.Println(email)
		}

		if pass == "" {
			pass2, err := READLINE.ReadPassword("Password: ")
			fmt.Println()

			if err != nil {
				if err != io.EOF && err != readline.ErrInterrupt {
					stdutil.PrintErr("Could not read password", err)
				}
				return
			}
			pass = string(pass2)
		}

		fmt.Println("Authenticating...")
		session, err = discordgo.New(email, pass)
	} else {
		fmt.Println("Authenticating...")
		if strings.HasPrefix(strings.ToLower(token), "user ") {
			token = token[len("user "):]
			USER = true
		} else {
			token = "Bot " + token
			USER = false
		}
		session, err = discordgo.New(token)
	}

	if err != nil {
		stdutil.PrintErr("Couldn't authenticate", err)
		return
	}

	user, err := session.User("@me")
	if err != nil {
		stdutil.PrintErr("Couldn't query user", err)
		return
	}

	ID = user.ID

	session.AddHandler(messageCreate)
	err = session.Open()
	if err != nil {
		stdutil.PrintErr("Could not open session", err)
	}

	fmt.Println("Logged in with user ID " + ID)
	fmt.Println("Write 'help' for help")
	fmt.Println("Press Ctrl+D or type 'exit' to exit.")

	for i := 0; i < 3; i++ {
		fmt.Println()
	}

	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		exit(session)
	}()

	COLOR_AUTOMATED.Set()

	if ar_lines != nil {
		for _, cmd := range ar_lines {
			cmd = strings.TrimSpace(cmd)
			if cmd == "" {
				continue
			}
			printPointer(session)
			fmt.Println(cmd)

			command(session, cmd)
		}
	}
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		printPointer(session)
		fmt.Println(cmd)

		command(session, cmd)
	}

	color.Unset()
	setCompleter(READLINE)

	for {
		COLOR_DEFAULT.Set()

		READLINE.SetPrompt(pointer(session))
		cmd, err := READLINE.Readline()

		color.Unset()

		if err != nil {
			if err != io.EOF && err != readline.ErrInterrupt {
				stdutil.PrintErr("Could not read line", err)
			} else {
				fmt.Println("exit")
			}
			exit(session)
			return
		}

		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		command(session, cmd)
	}
}

func exit(session *discordgo.Session) {
	color.Unset()
	playing = ""
	session.Close()
	os.Exit(0)
}

func execute(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printMessage(session *discordgo.Session, msg *discordgo.Message, prefixR bool, guild *discordgo.Guild, channel *discordgo.Channel) {
	var s string
	if prefixR {
		s += "\r"
	}
	s += "("

	if channel.IsPrivate {
		s += "Private"
	} else {
		s += guild.Name + " " + "#" + channel.Name
	}

	s += ") " + msg.Author.Username + ": " + msg.Content
	s += strings.Repeat(" ", 5)

	color.Unset()
	color.Yellow(s)
	COLOR_DEFAULT.Set()
}

func messageCreate(session *discordgo.Session, e *discordgo.MessageCreate) {
	channel, err := session.Channel(e.ChannelID)
	if err != nil {
		stdutil.PrintErr("Could not get channel", err)
		return
	}

	var guild *discordgo.Guild
	if !channel.IsPrivate {
		guild, err = session.Guild(channel.GuildID)
		if err != nil {
			stdutil.PrintErr("Could not get guild", err)
			return
		}
	}

	if messageCommand(session, e.Message, guild, channel) {
		return
	}

	lastMsg = location{
		guild:   guild,
		channel: channel,
	}

	hasOutput := false

	if messages {
		printMessage(session, e.Message, true, guild, channel)
		hasOutput = true
	}

	if len(luaMessageEvents) > 0 {
		hasOutput = true

		color.Unset()
		COLOR_AUTOMATED.Set()

		fmt.Print("\r" + strings.Repeat(" ", 20) + "\r")
		luaMessageEvent(session, e.Message)

		color.Unset()
		COLOR_DEFAULT.Set()
	}
	if hasOutput {
		printPointer(session)
	}
}

func messageCommand(session *discordgo.Session, e *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel) (isCmd bool) {
	if e.Author.ID != ID {
		return
	} else if !intercept {
		return
	}

	contents := strings.TrimSpace(e.Content)
	if !strings.HasPrefix(contents, "console.") {
		return
	}
	cmd := contents[len("console."):]

	isCmd = true

	if strings.EqualFold(cmd, "ping") {
		now := time.Now()

		_, err := session.ChannelMessageEdit(e.ChannelID, e.ID, "Pong!")
		if err != nil {
			stdutil.PrintErr("Couldn't edit message", err)
			return
		}

		now2 := time.Now()
		timestamp, err := e.Timestamp.Parse()
		if err != nil {
			stdutil.PrintErr("Couldn't parse timestamp", err)
			return
		}

		in := now.Sub(timestamp)
		out := now2.Sub(now)

		inMS := int(in.Nanoseconds() / time.Millisecond.Nanoseconds())
		outMS := int(out.Nanoseconds() / time.Millisecond.Nanoseconds())

		_, err = session.ChannelMessageEditEmbed(e.ChannelID, e.ID, &discordgo.MessageEmbed{
			Description: fmt.Sprintf("In: `%dms`\nOut: `%dms`", inMS, outMS),
		})
		if err != nil {
			stdutil.PrintErr("Couldn't edit message", err)
		}
		return
	}

	err := session.ChannelMessageDelete(e.ChannelID, e.ID)
	if err != nil {
		stdutil.PrintErr("Could not delete message", err)
	}

	lastLoc = loc
	loc = location{
		guild:   guild,
		channel: channel,
	}
	pointerCache = ""

	color.Unset()
	COLOR_AUTOMATED.Set()

	fmt.Println(cmd)
	command(session, cmd)

	color.Unset()
	COLOR_DEFAULT.Set()

	printPointer(session)
	return
}

const EMPTY_POINTER = "> "

var pointerCache string

func printPointer(session *discordgo.Session) {
	fmt.Print(pointer(session))
}
func pointer(session *discordgo.Session) string {
	if pointerCache != "" {
		return pointerCache
	}

	if loc.channel == nil {
		return EMPTY_POINTER
	}

	s := ""

	if loc.channel.IsPrivate {
		recipient := "Unknown"
		if loc.channel.Recipient != nil {
			recipient = loc.channel.Recipient.Username
		}
		s += "Private (" + recipient + ")"
	} else {
		guild := ""
		if loc.guild != nil {
			guild = loc.guild.Name
		}
		s += guild + " (#" + loc.channel.Name + ")"
	}

	s += EMPTY_POINTER
	pointerCache = s
	return s
}
