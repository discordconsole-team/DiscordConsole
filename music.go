/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2020 Mnpn

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
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/jD91mZM2/stdutil"
)

var vc *discordgo.VoiceConnection

var playing string
var cacheAudio = make(map[string][][]byte, 0)

var errDcaNegaitve = errors.New("negative number in DCA file")

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
			return errDcaNegaitve
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
	err := vc.Speaking(true)
	if err != nil {
		stdutil.PrintErr(tl("failed.voice.speak"), err)
		return
	}
	defer func() {
		err = vc.Speaking(false)
		if err != nil {
			stdutil.PrintErr(tl("failed.voice.speak"), err)
			return
		}
	}()

	for _, buf := range buffer {
		if playing == "" {
			break
		}
		vc.OpusSend <- buf
	}
}
