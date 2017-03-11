package main;

import "unicode";

func toEmojiString(text string) string{
	output := "";

	for _, c := range text{
		if(c >= 'a' && c <= 'z'){
			output += regional_indicator(c);
		} else if(c >= 'A' && c <= 'Z'){
			output += regional_indicator(unicode.ToLower(c));
		} else {
			switch c{
				case '-': output += ":heavy_minus_sign:";
				case '+': output += ":heavy_plus_sign:";
				case '$': output += ":heavy_dollar_sign:";
				case '*': output += ":heavy_asterisk_sign:";
				case '!': output += ":exclamation:";
				case '?': output += ":question:";
				case ' ': output += "\t";

				case '0': output += ":zero:";
				case '1': output += ":one:";
				case '2': output += ":two:";
				case '3': output += ":three:";
				case '4': output += ":four:";
				case '5': output += ":five:";
				case '6': output += ":six:";
				case '7': output += ":seven:";
				case '8': output += ":eight:";
				case '9': output += ":nine:";

				default: output += string(c);
			}
		}
	}

	return output;
}
func toEmoji(c rune) rune{
	if(c >= 'a' && c <= 'z'){
		return c - 'a' + '🇦';
	} else if(c >= 'A' && c <= 'Z'){
		return c - 'A' + '🇦';
	} else {
		switch c{
			case '-': return '➖';
			case '+': return '➕';
			case '$': return '💲';
			case '*': return '\u2731';
			case '!': return '❗';
			case '?': return '❓';
			default: return c;
		}
	}
}

func regional_indicator(c rune) string{
	return ":regional_indicator_" + string(c) + ":";
}
