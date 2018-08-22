package main

import (
	"testing"

	"github.com/fatih/color"
)

func TestAlias(t *testing.T) {
	// Alias known to use map
	go func() {
		command(nil, commandSource{
			//	NoMutex: true,
		}, "alias test exec echo hi", color.Output)
	}()
	go func() {
		command(nil, commandSource{
			//	NoMutex: true,
		}, "test", color.Output)
	}()
}
