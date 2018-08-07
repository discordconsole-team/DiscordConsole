/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2018 Mnpn

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/jD91mZM2/stdutil"
)

type apiData struct {
	Command string
	SentAt  int64
}

var apiTicker *time.Ticker
var apiDone = make(chan bool, 1)
var apiName = ""
var apiLast int64

func apiStart(session *discordgo.Session) (string, error) {
	if apiName != "" {
		return "", nil
	}
	f, err := ioutil.TempFile("", "DiscordConsole")
	if err != nil {
		return "", err
	}

	_, err = f.WriteString("{}")
	if err != nil {
		return "", err
	}

	name := f.Name()
	f.Close()

	go func(session *discordgo.Session, name string) {
		apiStartName(session, name)

		err := os.Remove(name)
		if err != nil {
			stdutil.PrintErr(tl("failed.file.delete")+" "+name, err)
		}
	}(session, name)
	return name, nil
}

func apiStartName(session *discordgo.Session, name string) {
	if apiName != "" {
		return
	}
	apiName = name
	apiTicker = time.NewTicker(time.Second * 2)
	for {
		select {
		case <-apiTicker.C:
			f, err := os.Open(name)
			if err != nil {
				stdutil.PrintErr(tl("failed.file.read")+" "+name, err)
				return
			}

			var data apiData

			err = json.NewDecoder(f).Decode(&data)
			f.Close()

			if err != nil {
				stdutil.PrintErr(tl("failed.json")+" "+name, err)
				continue
			}

			if data.SentAt == apiLast {
				continue
			}
			apiLast = data.SentAt

			cmd := data.Command
			if cmd == "" {
				continue
			}

			colorAutomated.Set()
			fmt.Println(cmd)
			command(session, commandSource{}, cmd, color.Output)

			color.Unset()
			colorDefault.Set()
			printPointer(session)
		case <-apiDone:
			return
		}
	}
}

func apiStop() {
	if apiTicker == nil {
		return
	}
	apiTicker.Stop()
	apiName = ""
	apiDone <- true
}

func apiSend(command string) error {
	if apiName == "" {
		return nil
	}

	api := apiData{
		Command: command,
		SentAt:  time.Now().Unix(),
	}
	apiLast = api.SentAt

	f, err := os.Create(apiName)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(api)
}
