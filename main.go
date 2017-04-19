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

	"github.com/bwmarrin/discordgo"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/legolord208/stdutil"
)

const autoRunFile = ".autorun"
const version = "2.1"

var devVersion = strings.Contains(version, "dev")

const (
	typeUser = iota
	typeBot
	typeWebhook
)

var closed bool

var userID string
var userToken string
var userType int

var rl *readline.Instance
var colorDefault = color.New(color.Bold)
var colorAutomated = color.New(color.Italic)
var colorMsg = color.New(color.FgYellow)
var colorError = color.New(color.FgRed, color.Bold)

const msgLimit = 2000

type stringArr []string

func (arr *stringArr) Set(val string) error {
	*arr = append(*arr, val)
	return nil
}

func (arr *stringArr) String() string {
	return "[" + strings.Join(*arr, " ") + "]"
}

func main() {
	defer handleCrash()

	var token string
	var email string
	var pass string
	var langfile string
	var help string
	var commands stringArr

	var noupdate bool
	var noautorun bool

	flag.StringVar(&token, "t", "", "Set token. Ignored if -e and/or -p are set.")
	flag.StringVar(&email, "e", "", "Set email.")
	flag.StringVar(&pass, "p", "", "Set password.")
	flag.StringVar(&langfile, "lang", "en", "Set language. Either a file path, or any of the following: en")
	flag.StringVar(&help, "lookup", "", "Search in `help` without starting the console")
	flag.Var(&commands, "x", "Pre-execute command. Can use flag multiple times.")

	flag.BoolVar(&noupdate, "noupdate", false, "Disable update checking.")
	flag.BoolVar(&noautorun, "noautorun", false, "Disable running commands in "+autoRunFile+" file.")
	flag.Parse()

	if help != "" {
		printHelp(help)
		return
	}

	doErrorHook()
	fmt.Println("DiscordConsole " + version)

	fmt.Println("Loading language...")
	switch langfile {
	case "en":
		loadLangDefault()
	case "sv":
		loadLangString(langSv)
	default:
		reader, err := os.Open(langfile)
		if err != nil {
			stdutil.PrintErr("Could not read language file", err)
			return
		}
		defer reader.Close()

		err = loadLang(reader)
		if err != nil {
			stdutil.PrintErr("Could not load language file", err)
			loadLangDefault()
		}
	}

	if !noupdate {
		fmt.Print(tl("update.checking") + " ")
		update, err := checkUpdate()
		if err != nil {
			stdutil.PrintErr(tl("update.error"), err)
		} else {
			if update.UpdateAvailable {
				fmt.Println()
				color.Cyan(tl("update.available") + " " + update.Version + ".")
				color.Cyan(tl("update.download") + " " + update.URL + ".")
			} else {
				fmt.Println(tl("update.none"))
			}
		}
	}

	fmt.Println(tl("loading.bookmarks"))
	err := loadBookmarks()
	if err != nil {
		stdutil.PrintErr(tl("failed.reading"), err)
	}

	var arLines []string
	if !noautorun {
		ar, err := ioutil.ReadFile(autoRunFile)
		if err != nil && os.IsExist(err) {
			stdutil.PrintErr(tl("failed.reading")+autoRunFile, err)
		} else if err == nil {
			arLines = strings.Split(string(ar), "\n")

			if len(arLines) > 0 {
				firstLine := arLines[0]
				if strings.HasPrefix(firstLine, ":") {
					token = firstLine[1:]
					arLines = arLines[1:]
				}
			}
		}
	}

	rl, err = readline.New(pointerEmpty)
	if err != nil {
		stdutil.PrintErr(tl("failed.realine.start"), err)
		return
	}

	fmt.Println()
	fmt.Println(tl("login.token"))
	fmt.Println(tl("login.token.user"))
	fmt.Println(tl("login.token.webhook"))
	fmt.Print("> ")
	if token == "" && email == "" && pass == "" {
		token, err = rl.Readline()
		if err != nil {
			if err != io.EOF && err != readline.ErrInterrupt {
				stdutil.PrintErr(tl("failed.realine.read"), err)
			}
			return
		}
	} else {
		if email != "" || pass != "" {
			token = ""
		}
		fmt.Println("[HIDDEN]")
	}

	var session *discordgo.Session
	if token == "" {
		userType = typeUser

		rl.SetPrompt("Email: ")
		if email == "" {
			email, err = rl.Readline()
		} else {
			fmt.Println(email)
		}

		if pass == "" {
			pass2, err := rl.ReadPassword("Password: ")
			fmt.Println()

			if err != nil {
				if err != io.EOF && err != readline.ErrInterrupt {
					stdutil.PrintErr(tl("failed.realine.read"), err)
				}
				return
			}
			pass = string(pass2)
		}

		fmt.Println(tl("login.starting"))
		session, err = discordgo.New(email, pass)
	} else {
		fmt.Println(tl("login.starting"))

		lower := strings.ToLower(token)

		if strings.HasPrefix(lower, "webhook ") {
			token = token[len("webhook "):]

			parts := strings.Split(token, "/")

			len := len(parts)
			if len >= 2 {
				userID = parts[len-2]
				userToken = parts[len-1]
			} else {
				stdutil.PrintErr(tl("invalid.webhook"), nil)
				return
			}

			userType = typeWebhook
			session, _ = discordgo.New(userToken)
		} else {
			if strings.HasPrefix(lower, "user ") {
				token = token[len("user "):]
				userType = typeUser
			} else {
				token = "Bot " + token
				userType = typeBot
				intercept = false
			}
			session, _ = discordgo.New(token)
		}
	}

	if userType != typeWebhook {
		if err != nil {
			stdutil.PrintErr(tl("failed.auth"), err)
			return
		}

		userToken = session.Token

		user, err := session.User("@me")
		if err != nil {
			stdutil.PrintErr(tl("failed.user"), err)
			return
		}

		userID = user.ID

		session.AddHandler(messageCreate)
		err = session.Open()
		if err != nil {
			stdutil.PrintErr(tl("failed.session.open"), err)
			return
		}

		fmt.Println(tl("login.finish") + " " + userID)
	}
	fmt.Println(tl("intro.help"))
	fmt.Println(tl("intro.exit"))

	for i := 0; i < 3; i++ {
		fmt.Println()
	}

	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		exit(session)
	}()

	colorAutomated.Set()

	if arLines != nil {
		for _, cmd := range arLines {
			printPointer(session)
			fmt.Println(cmd)

			command(session, cmd, color.Output)
		}
	}
	for _, cmd := range commands {
		printPointer(session)
		fmt.Println(cmd)

		command(session, cmd, color.Output)
	}

	color.Unset()
	setCompleter(rl)

	for {
		colorDefault.Set()

		rl.SetPrompt(pointer(session))
		cmd, err := rl.Readline()

		color.Unset()

		if err != nil {
			if err != io.EOF && err != readline.ErrInterrupt {
				stdutil.PrintErr(tl("failed.realine.read"), err)
			} else {
				fmt.Println("exit")
			}
			exit(session)
			return
		}

		command(session, cmd, color.Output)
		if closed {
			break
		}
	}
}

func exit(session *discordgo.Session) {
	closed = true

	apiStop()
	playing = ""

	if typeUser != typeWebhook {
		session.Close()
	}
	color.Unset()
}

func execute(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printMessage(session *discordgo.Session, msg *discordgo.Message, prefixR bool, guild *discordgo.Guild, channel *discordgo.Channel, w io.Writer) {
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
	colorMsg.Set()
	writeln(w, s)
	color.Unset()
	colorDefault.Set()
}

func writeln(w io.Writer, line string) error {
	// No error catching for now.
	// Because... if printing out fails,
	// chances are printing the error also fails
	_, err := w.Write([]byte(line + "\n"))
	return err
}

func handleCrash() {
	if val := recover(); val != nil {
		// No translations here. We wanna be as safe as possible
		stdutil.PrintErr("DiscordConsole has crashed.", nil)
		stdutil.PrintErr("Please tell LEGOlord208 what you did to cause this.", nil)
		stdutil.PrintErr("https://legolord208.github.io/contact", nil)
		stdutil.PrintErr("Error Details: "+fmt.Sprint(val), nil)
	}
}

const pointerEmpty = "> "

var pointerCache string

func printPointer(session *discordgo.Session) {
	fmt.Print(pointer(session))
}
func pointer(session *discordgo.Session) string {
	if pointerCache != "" {
		return pointerCache
	}

	if loc.channel == nil {
		return pointerEmpty
	}

	s := ""

	if loc.channel.IsPrivate {
		recipient := tl("pointer.unknown")
		if loc.channel.Recipient != nil {
			recipient = loc.channel.Recipient.Username
		}
		s += tl("pointer.private") + " (" + recipient + ")"
	} else {
		guild := ""
		if loc.guild != nil {
			guild = loc.guild.Name
		}
		s += guild + " (#" + loc.channel.Name + ")"
	}

	s += pointerEmpty
	pointerCache = s
	return s
}
