package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"time"
)

var ErrMsgNotFound = errors.New("Message not found!")

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
	if UserType == TypeUser {
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
		return nil, ErrMsgNotFound
	} else {
		return session.ChannelMessage(channel, msgID)
	}
}
