/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2017  LEGOlord208

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
	NoMutex  bool
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
var typeChannel = map[discordgo.ChannelType]string{
	discordgo.ChannelTypeDM:            "DM",
	discordgo.ChannelTypeGroupDM:       "Group",
	discordgo.ChannelTypeGuildCategory: "Category",
	discordgo.ChannelTypeGuildText:     "Text",
	discordgo.ChannelTypeGuildVoice:    "Voice",
}

const (
	messagesNone = iota
	messagesCurrent
	messagesPrivate
	messagesMentions
	messagesAll
)

type stringArr []string

func (arr *stringArr) Set(val string) error {
	*arr = append(*arr, val)
	return nil
}

func (arr *stringArr) String() string {
	return "[" + strings.Join(*arr, " ") + "]"
}
