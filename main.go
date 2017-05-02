/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2017  LEGOlord208

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
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
const version = "2.3.2"

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
var session *discordgo.Session

var rl *readline.Instance
var colorDefault = color.New(color.Bold)
var colorAutomated = color.New(color.Italic)
var colorMsg = color.New(color.FgYellow)
var colorChatMode = color.New(color.FgBlue)
var colorError = color.New(color.FgRed, color.Bold)

const msgLimit = 2000

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

	fmt.Println(`
DiscordConsole  Copyright (C) 2017  LEGOlord208
This program comes with ABSOLUTELY NO WARRANTY
This is free software, and you are welcome to redistribute it
under certain conditions.
`)

	loadLangAuto(langfile)

	if !noupdate {
		fmt.Print(tl("update.checking") + " ")
		update, err := checkUpdate()
		if err != nil {
			stdutil.PrintErr(tl("update.error"), err)
		} else {
			if update.UpdateAvailable {
				fmt.Println()
				color.Cyan(tl("update.available") + " " + update.Version)
				color.Cyan(tl("update.download") + " " + update.URL)
			} else {
				fmt.Println(tl("update.none"))
			}
		}
	}

	fmt.Println(tl("loading.bookmarks"))
	err := loadBookmarks()
	if err != nil {
		stdutil.PrintErr(tl("failed.file.read"), err)
	}

	var arLines []string
	if !noautorun {
		ar, err := ioutil.ReadFile(autoRunFile)
		if err != nil {
			if !os.IsNotExist(err) {
				stdutil.PrintErr(tl("failed.file.read")+autoRunFile, err)
			}
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
		stdutil.PrintErr(tl("failed.readline.start"), err)
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
				stdutil.PrintErr(tl("failed.readline.read"), err)
			}
			return
		}
	} else {
		if email != "" || pass != "" {
			token = ""
		}
		fmt.Println("[HIDDEN]")
	}

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
					stdutil.PrintErr(tl("failed.readline.read"), err)
				}
				return
			}
			pass = string(pass2)
		}

		fmt.Println(tl("login.starting"))
		session, err = discordgo.New(email, pass)
		if err == discordgo.ErrMFA {
			stdutil.PrintErr(tl("failed.mfa"), nil)
			return
		}
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
				if !strings.HasPrefix(token, "Bot ") {
					token = "Bot " + token
				}
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

		session.AddHandler(ready)
		session.AddHandler(guildCreate)
		session.AddHandler(guildDelete)
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

	for _, cmd := range arLines {
		if cmd == "" {
			continue
		}
		printPointer(session)
		fmt.Println(cmd)

		command(session, commandSource{Terminal: true}, cmd, color.Output)
	}
	for _, cmd := range commands {
		printPointer(session)
		fmt.Println(cmd)

		command(session, commandSource{Terminal: true}, cmd, color.Output)
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
				stdutil.PrintErr(tl("failed.readline.read"), err)
			} else {
				fmt.Println("exit")
			}
			exit(session)
			return
		}

		command(session, commandSource{Terminal: true}, cmd, color.Output)
		if closed {
			break
		}
	}
}

func exit(session *discordgo.Session) {
	closed = true

	apiStop()
	playing = ""
	if vc != nil {
		vc.Disconnect()
	}

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

func say(session *discordgo.Session, w io.Writer, channel, str string) (*discordgo.Message, bool) {
	if userType == typeWebhook {
		err := session.WebhookExecute(userID, userToken, false, &discordgo.WebhookParams{
			Content: str,
		})
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.send"), err)
			return nil, false
		}
		return nil, true
	}

	msg, err := session.ChannelMessageSend(loc.channel.ID, str)
	if err != nil {
		stdutil.PrintErr(tl("failed.msg.send"), err)
		return nil, false
	}
	writeln(w, tl("status.msg.create")+" "+msg.ID)

	return msg, true
}
