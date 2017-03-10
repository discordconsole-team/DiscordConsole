package main;

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"net/url"
	"encoding/json"
	"time"
)

func timestamp(e *discordgo.Message) (string, error){
	t, err := e.Timestamp.Parse();
	if(err != nil){
		return "", err;
	}

	s := t.Format(time.ANSIC);

	if(e.EditedTimestamp != ""){
		s += "*";
	}

	return s, nil;
}
func getMessage(session *discordgo.Session, channel, msgID string) (*discordgo.Message, error){
	if(USER){
		// Discord API does not allow getting specific message for users.
		// DiscordGo **stable** does not support the "around" setting.
		// Workaround? Manually

		//msgs, err = session.ChannelMessages(loc.channelID, 3, "", "", msgID);
		v := url.Values{};
		v.Set("limit", "3");
		v.Set("around", msgID);

		endpoint := discordgo.EndpointChannelMessages(channel);
		body, err := session.RequestWithBucketID("GET", endpoint + "?" + v.Encode(), nil, endpoint);
		if(err != nil){
			return nil, err;
		}

		var msgs []*discordgo.Message;
		err = json.Unmarshal(body, &msgs);
		if(err != nil){
			return nil, err;
		}

		for _, m := range msgs{
			if(m.ID == msgID){
				return m, nil;
			}
		}
		return nil, errors.New("Message not found!");
	} else {
		return session.ChannelMessage(channel, msgID);
	}
}
