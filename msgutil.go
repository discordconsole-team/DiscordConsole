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
