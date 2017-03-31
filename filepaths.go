package main

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func fixPath(path *string) error {
	s := *path

	for {
		i := strings.Index(s, "~")
		if i < 0 {
			break
		}

		current, err := user.Current()
		if err != nil {
			return errors.New("Could not determine value of ~, " + err.Error())
		}

		s = filepath.Join(s[:i], current.HomeDir, s[i+1:])
	}
	for _, env := range os.Environ() {
		keyval := strings.SplitN(env, "=", -1)

		s = strings.Replace(s, "%"+keyval[0]+"%", keyval[1], -1)
		s = strings.Replace(s, "$"+keyval[0], keyval[1], -1)
	}

	*path = s
	return nil
}
