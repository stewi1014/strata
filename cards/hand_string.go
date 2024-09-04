package cards

import "strings"

// String returns the name of the hand, and the rankers in brackets if present.
// e.g. a 6-triple, Ace, King three of a kind hand would return Three Of A Kind (6, A, K)
func (h Hand) String() string {
	if h == royalFlushHand {
		return "Royal Flush"
	}

	if h&^handKindMask == 0 {
		return h.Name()
	}

	var str strings.Builder
	str.WriteString(h.Name())

	str.WriteString(" (")
	ranks := h.Rankers()
	for i, rank := range ranks {
		str.WriteRune(rank.RankCharacter())

		if i < len(ranks)-1 {
			str.WriteString(", ")
		}
	}
	str.WriteString(")")

	return str.String()
}

// Name returns the name of the hand with no other informtion,
// e.g. 6-triple, 9-pair full house would return "Full House".
func (h Hand) Name() string {
	switch h & handKindMask {
	case StraightFlush:
		return "Straight Flush"
	case FourOfAKind:
		return "Four Of A Kind"
	case FullHouse:
		return "Full House"
	case Flush:
		return "Flush"
	case Straight:
		return "Straight"
	case ThreeOfAKind:
		return "Three Of A Kind"
	case TwoPair:
		return "Two Pair"
	case OnePair:
		return "One Pair"
	case HighCard:
		return "High Card"
	default:
		return "(invalid hand)"
	}
}
