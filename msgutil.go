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
	"time"

	"github.com/bwmarrin/discordgo"
)

var errMsgNotFound = errors.New("message not found")

func timestamp(e *discordgo.Message) (string, error) {
	t, err := e.Timestamp.Parse()
	if err != nil {
		return "", err
	}

	s := t.Format(time.ANSIC)

	if e.EditedTimestamp != "" {
		s += "*"
	}

	return s, nil
}
func getMessage(session *discordgo.Session, channel, msgID string) (*discordgo.Message, error) {
	if userType == typeUser {
		msgs, err := session.ChannelMessages(channel, 3, "", "", msgID)
		if err != nil {
			return nil, err
		}

		// This comment is a gravestone as memory
		// from when I did a web request manually
		// because I didn't wanna use the develop branch.
		// lol.

		for _, m := range msgs {
			if m.ID == msgID {
				return m, nil
			}
		}
		return nil, errMsgNotFound
	}
	return session.ChannelMessage(channel, msgID)
}
