package main

import (
	"github.com/fatih/color"
	"github.com/legolord208/stdutil"
	"os"
	"strings"
)

func doHook() {
	stdutil.EventPrePrintError = append(stdutil.EventPrePrintError, func(full string, msg string, err error) bool {
		if err != nil && isPermission(err) {
			COLOR_ERROR.Fprintln(os.Stderr, "No permissions to perform this action.")
			return true
		}
		color.Unset()
		COLOR_ERROR.Set()
		return false
	})
	stdutil.EventPostPrintError = append(stdutil.EventPostPrintError, func(text string, msg string, err error) {
		color.Unset()
	})
}

func isPermission(err error) bool {
	return strings.Contains(err.Error(), "Missing Permission")
}
