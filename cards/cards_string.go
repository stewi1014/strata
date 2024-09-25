package cards

import (
	"errors"
	"math/bits"
	"strings"
	"unicode"
)

var (
	ErrInvalidFormat = errors.New("invalid format")
)

// ParseCards finds all instances of a valid rank character followed
// by a valid suite character.
//
// e.g.
//
//	"AS qd" = AceSpades | QueenDiamonds
//	"this is actually valid" = TenHearts | AceClubs ('th' in 'this', and 'ac' in 'actually)
func ParseCards(str string) (cards Cards) {
	r := strings.NewReader(str)

	for {
		// get rank
		char, _, err := r.ReadRune()
		if err != nil {
			return
		}

		rank, err := RankFromCharacter(char)
		if err != nil {
			continue
		}

		// we have a valid rank
		// get suite
		char, _, err = r.ReadRune()
		if err != nil {
			return
		}

		suite, err := SuiteFromCharacter(char)
		if err != nil {
			continue
		}

		cards |= (rank & suite)
		if cards == Deck {
			// no point continuing, every card already exists.
			return cards
		}
	}
}

func (a Cards) RankCharacter() rune {
	return []rune{
		'\uFFFD',
		'\uFFFD',
		'2',
		'3',
		'4',
		'5',
		'6',
		'7',
		'8',
		'9',
		'T',
		'J',
		'Q',
		'K',
		'A',
	}[a.RankAsInt()]
}

func RankFromCharacter(char rune) (Cards, error) {
	switch unicode.ToUpper(char) {
	case '2':
		return Two, nil
	case '3':
		return Three, nil
	case '4':
		return Four, nil
	case '5':
		return Five, nil
	case '6':
		return Six, nil
	case '7':
		return Seven, nil
	case '8':
		return Eight, nil
	case '9':
		return Nine, nil
	case 'T':
		return Ten, nil
	case 'J':
		return Jack, nil
	case 'Q':
		return Queen, nil
	case 'K':
		return King, nil
	case 'A':
		return Ace, nil
	}

	return 0, ErrInvalidRank
}

func (a Cards) RankString() string {
	return []string{
		"\uFFFD",
		"\uFFFD",
		"Two",
		"Three",
		"Four",
		"Five",
		"Six",
		"Seven",
		"Eight",
		"Nine",
		"Ten",
		"Jack",
		"Queen",
		"King",
		"Ace",
	}[a.RankAsInt()]
}

func (a Cards) SuiteIcon() rune {
	return []rune{
		'\uFFFD',
		'♧',
		'♢',
		'♡',
		'♤',
	}[a.SuiteAsInt()]
}

func (a Cards) SuiteCharacter() rune {
	return []rune{
		'\uFFFD',
		'c',
		'd',
		'h',
		's',
	}[a.SuiteAsInt()]
}

func SuiteFromCharacter(char rune) (Cards, error) {
	switch unicode.ToLower(char) {
	case 'c':
		return Clubs, nil
	case 'd':
		return Diamonds, nil
	case 'h':
		return Hearts, nil
	case 's':
		return Spades, nil
	}

	return 0, ErrInvalidSuite
}

func (a Cards) SuiteString() string {
	return []string{
		"\uFFFD",
		"Club",
		"Diamond",
		"Heart",
		"Spade",
	}[a.SuiteAsInt()]
}

func (a Cards) Icon() rune {
	return []rune{
		'🃒', '🃂', '🂲', '🂢',
		'🃓', '🃃', '🂳', '🂣',
		'🃔', '🃄', '🂴', '🂤',
		'🃕', '🃅', '🂵', '🂥',
		'🃖', '🃆', '🂶', '🂦',
		'🃗', '🃇', '🂷', '🂧',
		'🃘', '🃈', '🂸', '🂨',
		'🃙', '🃉', '🂹', '🂩',
		'🃚', '🃊', '🂺', '🂪',
		'🃛', '🃋', '🂻', '🂫',
		'🃝', '🃍', '🂽', '🂭',
		'🃞', '🃎', '🂾', '🂮',
		'🃑', '🃁', '🂱', '🂡',
	}[bits.TrailingZeros64(uint64(a&Deck))]
}

// String returns the full names (e.g. "Ace of Spades, Nine of Diamonds") if 7 or less cards exist.
// If there is more than 7 cards, it uses abbreviated form (e.g. "As, Ks, Qs, Js, 10s, 9d, 7h, 5d, 4c")
func (a Cards) String() string {
	a &= Deck // Ignore 12MSBs

	if a.Len() == 0 {
		return "\uFFFD"
	}

	if a.Len() <= 7 {
		i := a.High()
		str := i.RankString() + " of " + i.SuiteString()
		for i = a.NextHigh(i); i > 0; i = a.NextHigh(i) {
			str += ", " + i.RankString() + " of " + i.SuiteString()
		}
		return str
	}

	i := a.High()
	str := string(i.RankCharacter()) + string(i.SuiteCharacter())
	for i = a.NextHigh(i); i > 0; i = a.NextHigh(i) {
		str += ", " + string(i.RankCharacter()) + string(i.SuiteCharacter())
	}

	return str
}

// Icons returns the icons of the cards (e.g. 🃘, 🂳)
func (a Cards) Icons() string {
	var buff strings.Builder

	i := a.High()
	buff.WriteRune(i.Icon())
	for i = a.NextHigh(i); i != 0; i = a.NextHigh(i) {
		buff.WriteString(" ")
		buff.WriteRune(i.Icon())
	}

	return buff.String()
}
