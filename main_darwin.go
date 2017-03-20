package main;

import (
	"os/user"
	"filepath"
)

func tokenDir() (string, error){
	current, err := user.Current();
	if(err != nil){
		return "", err;
	}
	return filepath.Join(current.HomeDir, "Library", "Application Support", "discord"), nil;
}
