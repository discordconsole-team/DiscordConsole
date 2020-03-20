package main

import (
	"errors"
	"io"
	"strings"
)

type tokenKind int

const (
	tokenSeparator tokenKind = iota
	tokenString
	tokenSubstitute
)

type token struct {
	kind tokenKind
	text string
}

type tokenizer struct {
	quote bool
}

func (tokenizer *tokenizer) nextToken(r *strings.Reader) (t token, err error) {
	text := ""
	for {
		var c rune
		c, _, err = r.ReadRune()
		if err == io.EOF && len(text) > 0 {
			t = token{
				kind: tokenString,
				text: text,
			}
			err = nil
			return
		}
		if err != nil {
			return
		}

		if c == '"' {
			tokenizer.quote = !tokenizer.quote
		} else if c == '$' {
			if len(text) > 0 {
				r.UnreadRune()
				t = token{
					kind: tokenString,
					text: text,
				}
				return
			}

			c, _, err = r.ReadRune()
			if err != nil {
				return
			}

			text += string(c)

			if c == '{' {
				text = ""
				for c != '}' {
					c, _, err = r.ReadRune()
					if err != nil {
						return
					}
					text += string(c)
				}
				text = text[:len(text)-1]
				t = token{
					kind: tokenSubstitute,
					text: text,
				}
				return
			}

			text += string(c)
		} else if c == '\\' {
			c, _, err = r.ReadRune()
			if err != nil {
				return
			}

			if c == 'n' {
				c = '\n'
			}

			text += string(c)
		} else if (c == ' ' || c == '\t') && !tokenizer.quote {
			if len(text) > 0 {
				r.UnreadByte()
				t = token{
					kind: tokenString,
					text: text,
				}
				return
			}

			t = token{
				kind: tokenSeparator,
				text: string(c),
			}
			return
		} else {
			text += string(c)
		}
	}
}

func parse(subst func(string) (string, bool), text string) ([]string, error) {
	reader := strings.NewReader(text)

	var previous []string
	current := ""

	var t tokenizer

	for {
		token, err := t.nextToken(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch token.kind {
		case tokenSeparator:
			previous = append(previous, current)
			current = ""
		case tokenSubstitute:
			val, ok := subst(token.text)
			if !ok {
				return nil, errors.New(tl("invalid.substitute") + ": " + token.text)
			}
			current += val
		case tokenString:
			current += token.text
		default:
			panic("Unreachable code: All tokens should be checked")
		}
	}

	if t.quote {
		return nil, errors.New(tl("invalid.unmatched.quote"))
	}
	previous = append(previous, current)

	return previous, nil
}
