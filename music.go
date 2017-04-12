package main

import (
	"encoding/binary"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/stdutil"
	"io"
	"os"
)

var playing string
var cacheAudio = make(map[string][][]byte, 0)

var ErrDcaNegaitve = errors.New("Negative number in DCA file")

func loadAudio(file string, buffer *[][]byte) error {
	cache, ok := cacheAudio[file]
	if ok {
		*buffer = cache
		return nil
	}

	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	var length int16
	for {
		err := binary.Read(reader, binary.LittleEndian, &length)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		} else if err != nil {
			return err
		}

		if length <= 0 {
			return ErrDcaNegaitve
		}

		buf := make([]byte, length)
		err = binary.Read(reader, binary.LittleEndian, buf)
		if err != nil {
			return err
		}

		*buffer = append(*buffer, buf)
	}

	cacheAudio[file] = *buffer
	return nil
}

func play(buffer [][]byte, session *discordgo.Session, guild, channel string) {
	vc, err := session.ChannelVoiceJoin(guild, channel, false, true)
	if err != nil {
		stdutil.PrintErr(tl("failed.voice.connect"), err)
		return
	}

	err = vc.Speaking(true)
	if err != nil {
		stdutil.PrintErr(tl("failed.voice.speak"), err)

		err = vc.Disconnect()
		if err != nil {
			stdutil.PrintErr(tl("failed.voice.disconnect"), err)
		}
		return
	}

	for _, buf := range buffer {
		if playing == "" {
			break
		}
		vc.OpusSend <- buf
	}

	err = vc.Speaking(false)
	if err != nil {
		stdutil.PrintErr(tl("failed.voice.speak"), err)
	}

	err = vc.Disconnect()
	if err != nil {
		stdutil.PrintErr(tl("failed.voice.disconnect"), err)
	}
}
