/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2017 Mnpn

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
	"errors"
	"io/ioutil"
	"net/http"
)

const updateURL = "https://api.github.com/repos/discordconsole-team/DiscordConsole/releases"

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
