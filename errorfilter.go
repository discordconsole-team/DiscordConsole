package main

import (
	"github.com/fatih/color"
	"github.com/legolord208/stdutil"
	"os"
	"strings"
)

func doErrorHook() {
	stdutil.EventPrePrintError = append(stdutil.EventPrePrintError, func(full string, msg string, err error) bool {
		if err != nil && isPermission(err) {
			ColorError.Fprintln(os.Stderr, lang["failed.perms"])
			return true
		}
		color.Unset()
		ColorError.Set()
		return false
	})
	stdutil.EventPostPrintError = append(stdutil.EventPostPrintError, func(text string, msg string, err error) {
		color.Unset()
	})
}

func isPermission(err error) bool {
	return strings.Contains(err.Error(), lang["failed.perms"])
}
