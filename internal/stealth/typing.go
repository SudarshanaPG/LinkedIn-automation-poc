package stealth

import (
	"context"
	"strings"
	"time"
	"unicode"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

var neighborKeys = map[rune]string{
	'a': "qwsz",
	'b': "vghn",
	'c': "xdfv",
	'd': "erfcxs",
	'e': "wsdr",
	'f': "rtgvcd",
	'g': "tyhbvf",
	'h': "yujnbg",
	'i': "ujko",
	'j': "uikmnh",
	'k': "iolmj",
	'l': "opk",
	'm': "njk",
	'n': "bhjm",
	'o': "iklp",
	'p': "ol",
	'q': "wa",
	'r': "edft",
	's': "wedxza",
	't': "rfgy",
	'u': "yhji",
	'v': "cfgb",
	'w': "qase",
	'x': "zsdc",
	'y': "tghu",
	'z': "asx",
}

func (c *Controller) TypeHuman(ctx context.Context, page *rod.Page, text string) error {
	for _, ch := range text {
		if err := ctx.Err(); err != nil {
			return err
		}
		if c.ShouldTypo() && isTypoCandidate(ch) {
			typo := c.randomNeighbor(ch)
			if typo != 0 {
				if err := page.InsertText(string(typo)); err != nil {
					return err
				}
				time.Sleep(c.TypingDelay())
				if err := page.Keyboard.Type(input.Backspace); err != nil {
					return err
				}
				time.Sleep(time.Duration(c.randomInt(30, 80)) * time.Millisecond)
			}
		}
		switch ch {
		case '\n':
			if err := page.Keyboard.Type(input.Enter); err != nil {
				return err
			}
		case '\t':
			if err := page.Keyboard.Type(input.Tab); err != nil {
				return err
			}
		default:
			if err := page.InsertText(string(ch)); err != nil {
				return err
			}
		}
		time.Sleep(c.TypingDelay())
	}
	return nil
}

func isTypoCandidate(ch rune) bool {
	return unicode.IsLetter(ch)
}

func (c *Controller) randomNeighbor(ch rune) rune {
	lower := unicode.ToLower(ch)
	candidates := neighborKeys[lower]
	if candidates == "" {
		return 0
	}
	selected := candidates[c.rng.Intn(len(candidates))]
	if unicode.IsUpper(ch) {
		return rune(strings.ToUpper(string(selected))[0])
	}
	return rune(selected)
}
