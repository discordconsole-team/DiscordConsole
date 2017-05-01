package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type keyval struct {
	Key string
	Val string
}

func (data *keyval) String() string {
	return data.Key + ": " + data.Val
}

func findValByKey(keyvals []*keyval, key string) (string, bool) {
	for _, keyval := range keyvals {
		if strings.EqualFold(key, keyval.Key) {
			return keyval.Val, true
		}
	}
	return "", false
}

type commandSource struct {
	Terminal bool
	Alias    bool
}

var typeRelationships = map[int]string{
	1: "Friend",
	2: "Blocked",
	3: "Incoming request",
	4: "Sent request",
}
var typeVerifications = map[discordgo.VerificationLevel]string{
	discordgo.VerificationLevelNone:   "None",
	discordgo.VerificationLevelLow:    "Low",
	discordgo.VerificationLevelMedium: "Medium",
	discordgo.VerificationLevelHigh:   "High",
}
var typeMessages = map[string]int{
	"all":      messagesAll,
	"mentions": messagesMentions,
	"private":  messagesPrivate,
	"current":  messagesCurrent,
	"none":     messagesNone,
}
var typeStatuses = map[string]discordgo.Status{
	"online":    discordgo.StatusOnline,
	"idle":      discordgo.StatusIdle,
	"dnd":       discordgo.StatusDoNotDisturb,
	"invisible": discordgo.StatusInvisible,
}

type stringArr []string

func (arr *stringArr) Set(val string) error {
	*arr = append(*arr, val)
	return nil
}

func (arr *stringArr) String() string {
	return "[" + strings.Join(*arr, " ") + "]"
}
