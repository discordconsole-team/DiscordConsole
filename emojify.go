/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2017 Mnpn

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

import "unicode"

func toEmojiString(c rune) string {
	if c >= 'a' && c <= 'z' {
		return regionalIndicator(c)
	} else if c >= 'A' && c <= 'Z' {
		return regionalIndicator(unicode.ToLower(c))
	} else {
		switch c {
		case '-':
			return ":heavy_minus_sign:"
		case '+':
			return ":heavy_plus_sign:"
		case '$':
			return ":heavy_dollar_sign:"
		case '*':
			return ":asterisk:"
		case '!':
			return ":exclamation:"
		case '?':
			return ":question:"
		case ' ':
			return "\t"

		case '0':
			return ":zero:"
		case '1':
			return ":one:"
		case '2':
			return ":two:"
		case '3':
			return ":three:"
		case '4':
			return ":four:"
		case '5':
			return ":five:"
		case '6':
			return ":six:"
		case '7':
			return ":seven:"
		case '8':
			return ":eight:"
		case '9':
			return ":nine:"

		default:
			return string(c)
		}
	}
}
func toEmoji(c rune) string {
	if c >= 'A' && c <= 'Z' {
		return string(c - 'A' + '🇦')
	}
	if c >= 'a' && c <= 'z' {
		return string(c - 'a' + '🇦')
	}
	if c >= '0' && c <= '9' || c == '*' {
		return string(c) + "\u20E3"
	}
	switch c {
	case '-':
		return "➖"
	case '+':
		return "➕"
	case '$':
		return "💲"
	case '!':
		return "❗"
	case '?':
		return "❓"
	default:
		return string(c)
	}
}

func regionalIndicator(c rune) string {
	return ":regional_indicator_" + string(c) + ":"
}
