package main

import (
	"encoding/json"
	"os"
)

const fileBookmarks = ".bookmarks"

var bookmarks = make(map[string]string)
var bookmarksCache = make(map[string]location)

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
