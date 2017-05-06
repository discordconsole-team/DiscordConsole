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
	"encoding/json"
	"os"
)

const fileBookmarks = ".bookmarks"

var bookmarks = make(map[string]string)
var bookmarksCache = make(map[string]*location)

func loadBookmarks() error {
	reader, err := os.Open(fileBookmarks)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer reader.Close()

	return json.NewDecoder(reader).Decode(&bookmarks)
}

func saveBookmarks() error {
	writer, err := os.Create(fileBookmarks)
	if err != nil {
		return err
	}
	defer writer.Close()

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	return encoder.Encode(bookmarks)
}
