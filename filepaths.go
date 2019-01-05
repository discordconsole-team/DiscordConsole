/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2019 Mnpn

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
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
			return errors.New(tl("failed.path.home") + ", " + err.Error())
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
