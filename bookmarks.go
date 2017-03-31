package main

import (
	"encoding/json"
	"github.com/legolord208/stdutil"
	"io/ioutil"
	"os"
)

type bookmark struct {
	GuildID   string
	ChannelID string
}

const BOOKMARKS_FILE = ".bookmarks"

var bookmarks = make(map[string]string)

func loadBookmarks() error {
	contents, err := ioutil.ReadFile(BOOKMARKS_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	//TODO return json.Unmarshal(contents, &bookmarks);

	err = json.Unmarshal(contents, &bookmarks)
	if err != nil {
		bookmarks2 := make(map[string]bookmark)

		err = json.Unmarshal(contents, &bookmarks2)
		if err != nil {
			return err
		}

		stdutil.PrintErr("Warning: An old bookmark system is used. This system will be unsupported in a later version.", nil)
		stdutil.PrintErr("You're highly suggested to edit any bookmark and back again for the new system to take effect.", nil)

		bookmarks = make(map[string]string, len(bookmarks2))

		for i, mark := range bookmarks2 {
			bookmarks[i] = mark.ChannelID
		}

	}
	return nil
}

func saveBookmarks() error {
	contents, err := json.MarshalIndent(bookmarks, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(BOOKMARKS_FILE, contents, 0666)
}
