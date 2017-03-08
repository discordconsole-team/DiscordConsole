package main;

import (
	"io/ioutil"
	"encoding/json"
	"os"
)

const BOOKMARKS_FILE = ".bookmarks";
var bookmarks = make(map[string]location)

func loadBookmarks() error{
	contents, err := ioutil.ReadFile(BOOKMARKS_FILE);
	if(err != nil){
		if(os.IsNotExist(err)){
			return nil;
		} else {
			return err;
		}
	}

	return json.Unmarshal(contents, &bookmarks);
}

func saveBookmarks() error{
	contents, err := json.MarshalIndent(bookmarks, "", "\t");
	if(err != nil){
		return err;
	}

	return ioutil.WriteFile(BOOKMARKS_FILE, contents, 0666);
}
