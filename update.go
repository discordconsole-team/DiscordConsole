package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

const UPDATE_URL = "https://api.github.com/repos/LEGOlord208/DiscordConsole/releases"

var ErrNoRelease = errors.New("No release available.")

type updateObj struct {
	Version         string `json:"tag_name"`
	Url             string `json:"html_url"`
	UpdateAvailable bool
}

func checkUpdate() (*updateObj, error) {
	res, err := http.Get(UPDATE_URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var updates []updateObj
	err = json.Unmarshal(content, &updates)
	if err != nil {
		return nil, err
	}

	if len(updates) < 1 {
		return nil, ErrNoRelease
	}

	update := updates[0]

	update.UpdateAvailable = update.Version != Version
	return &update, nil
}
