/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2018 Mnpn

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
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jD91mZM2/stdutil"
	"github.com/mattn/go-colorable"
)

func doErrorHook() {
	stdutil.EventPrePrintError = append(stdutil.EventPrePrintError, func(full string, msg string, err error) bool {
		if err != nil && isPermission(err) {
			colorError.Fprintln(colorable.NewColorable(os.Stderr), tl("failed.perms"))
			return true
		}
		color.Unset()
		colorError.Set()
		return false
	})
	stdutil.EventPostPrintError = append(stdutil.EventPostPrintError, func(text string, msg string, err error) {
		color.Unset()
	})
}

func isPermission(err error) bool {
	return strings.Contains(err.Error(), "Missing Permission")
}
