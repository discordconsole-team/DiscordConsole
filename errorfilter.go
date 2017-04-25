package main

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/legolord208/stdutil"
)

func doErrorHook() {
	stdutil.EventPrePrintError = append(stdutil.EventPrePrintError, func(full string, msg string, err error) bool {
		if err != nil && isPermission(err) {
			colorError.Fprintln(os.Stderr, tl("failed.perms"))
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
	return strings.Contains(err.Error(), "permission")
}
