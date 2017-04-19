package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

const updateURL = "https://api.github.com/repos/LEGOlord208/DiscordConsole/releases"

var errNoRelease = errors.New("no release available")

type updateObj struct {
	Version         string `json:"tag_name"`
	URL             string `json:"html_url"`
	UpdateAvailable bool
}

func checkUpdate() (*updateObj, error) {
	res, err := http.Get(updateURL)
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
		return nil, errNoRelease
	}

	update := updates[0]

	update.UpdateAvailable = update.Version != version
	return &update, nil
}
