package main;

import (
	"os/user"
	"path/filepath"
)

func tokenDir() (string, error){
	current, err := user.Current();
	if(err != nil){
		return "", err;
	}
	return filepath.Join(current.HomeDir, ".config", "discord"), nil;
}
