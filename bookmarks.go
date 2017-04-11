package main

import (
	"encoding/json"
	"os"
)

type bookmark struct {
	GuildID   string
	ChannelID string
}

const FileBookmarks = ".bookmarks"

var bookmarks = make(map[string]string)

func loadBookmarks() error {
	reader, err := os.Open(FileBookmarks)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	defer reader.Close()

	return json.NewDecoder(reader).Decode(&bookmarks)
}

func saveBookmarks() error {
	writer, err := os.Open(FileBookmarks)
	if err != nil {
		return err
	}
	defer writer.Close()

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	return encoder.Encode(bookmarks)
}
