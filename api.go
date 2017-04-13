package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/legolord208/stdutil"
)

type apiData struct {
	Command string
	SentAt  int64
}

var api_ticker *time.Ticker
var api_done = make(chan bool, 1)
var api_name = ""
var api_last int64

func api_start(session *discordgo.Session) (string, error) {
	if api_name != "" {
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
		api_start_name(session, name)

		fmt.Println("removing file")
		err := os.Remove(name)
		if err != nil {
			stdutil.PrintErr(tl("failed.file.delete")+" "+name, err)
		}
		fmt.Println("u woot")
	}(session, name)
	return name, nil
}

func api_start_name(session *discordgo.Session, name string) {
	if api_name != "" {
		return
	}
	api_name = name
	api_ticker = time.NewTicker(time.Second * 2)
	for {
		select {
		case <-api_ticker.C:
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

			if data.SentAt == api_last {
				continue
			}
			api_last = data.SentAt

			cmd := data.Command
			if cmd == "" {
				continue
			}

			ColorAutomated.Set()
			fmt.Println(cmd)
			command(session, cmd)

			color.Unset()
			ColorDefault.Set()
			printPointer(session)
		case <-api_done:
			return
		}
	}
}

func api_stop() {
	if api_ticker == nil {
		return
	}
	api_ticker.Stop()
	api_name = ""
	api_done <- true
}

func api_send(command string) error {
	if api_name == "" {
		return nil
	}

	api := apiData{
		Command: command,
		SentAt:  time.Now().Unix(),
	}
	api_last = api.SentAt

	f, err := os.Create(api_name)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(api)
}
