package main;

import "unicode";

func toEmojiString(c rune) string{
	if(c >= 'a' && c <= 'z'){
		return regional_indicator(c);
	} else if(c >= 'A' && c <= 'Z'){
		return regional_indicator(unicode.ToLower(c));
	} else {
		switch c{
			case '-': return ":heavy_minus_sign:";
			case '+': return ":heavy_plus_sign:";
			case '$': return ":heavy_dollar_sign:";
			case '*': return ":heavy_asterisk_sign:";
			case '!': return ":exclamation:";
			case '?': return ":question:";
			case ' ': return "\t";

			case '0': return ":zero:";
			case '1': return ":one:";
			case '2': return ":two:";
			case '3': return ":three:";
			case '4': return ":four:";
			case '5': return ":five:";
			case '6': return ":six:";
			case '7': return ":seven:";
			case '8': return ":eight:";
			case '9': return ":nine:";

			default: return string(c);
		}
	}
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
