package main;

import (
	"os"
	"io"
	"encoding/binary"
	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/stdutil"
)

var playing string;
var cacheAudio = make(map[string][][]byte, 0);

func load(file string, buffer* [][]byte) error{
	cache, ok := cacheAudio[file];
	if(ok){
		*buffer = cache;
		return nil;
	}

	reader, err := os.Open(file);
	if(err != nil){
		return err;
	}
	defer reader.Close();

	var length int16;
	for{
		err := binary.Read(reader, binary.LittleEndian, &length);
		if(err == io.EOF || err == io.ErrUnexpectedEOF){
			break;
		} else if(err != nil){
			return err;
		}

		buf := make([]byte, length);
		err = binary.Read(reader, binary.LittleEndian, buf);
		if(err != nil){
			return err;
		}

		*buffer = append(*buffer, buf);
	}

	cacheAudio[file] = *buffer;
	return nil;
}

func play(buffer [][]byte, session *discordgo.Session, guild, channel string){
	vc, err := session.ChannelVoiceJoin(guild, channel, false, true);
	if(err != nil){
		stdutil.PrintErr("Could not connect to voice channel", err);
		return;
	}

	err = vc.Speaking(true);
	if(err != nil){
		stdutil.PrintErr("Could not start speaking", err);

		err = vc.Disconnect();
		if(err != nil){
			stdutil.PrintErr("Could not disconnect", err);
		}
		return;
	}

	for _, buf := range buffer{
		if(playing == ""){ break; }
		vc.OpusSend <- buf;
	}

	err = vc.Speaking(false);
	if(err != nil){
		stdutil.PrintErr("Could not stop speaking", err);
	}

	err = vc.Disconnect();
	if(err != nil){
		stdutil.PrintErr("Could not disconnect", err);
	}
}
