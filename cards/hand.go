package cards

import (
	"fmt"
)

const (
	handKindMask    = Hand(0xF0000000)
	royalFlushCards = AceSpades | KingSpades | QueenSpades | JackSpades | TenSpades

	StraightFlush = Hand(0x90000000) // An undefined Straight Flush.
	FourOfAKind   = Hand(0x80000000) // An indefined Four of a Kind.
	FullHouse     = Hand(0x70000000) // An undefined Full House.
	Flush         = Hand(0x60000000) // An undefined Flush.
	Straight      = Hand(0x50000000) // An undefined Straight.
	ThreeOfAKind  = Hand(0x40000000) // An undefined Three of a kind.
	TwoPair       = Hand(0x30000000) // An undefined Two Pair.
	OnePair       = Hand(0x20000000) // An undefined Pair.
	HighCard      = Hand(0x10000000) // An undefined High Card.
)

var royalFlushHand = NewHand(StraightFlush, Ace)
var handRankers = []int{
	0, // invalid
	5, // High card
	4, // One Pair
	3, // Two Pair
	3, // Three of a Kind
	1, // Straight
	5, // Flush
	2, // Full house
	2, // Four of a Kind
	1, // Straight Flush
}

// Hand is a comparable integer representation of a [poker hand].
//
// Integer comparason operators (>, < and =) are the intended method for comparing hands.
//
// The most significant 4 bits store the kind of hand (e.g. Flush, TwoPair etc...),
// with each following 4 bits storing the relevent ranks.
//
// [poker hand]: https://en.wikipedia.org/wiki/List_of_poker_hands
type Hand uint32

// NewHand returns a new hand given the kind of hand and rankers required to rank the hand.
// This method does not translate a set of cards into a poker hand,
// instead this method provides a way to populate hand ranking infomration.
// To find the best hand from a given set of cards, see GetHand().
//
// hand is the kind of hand to be created,
// it panics if hand is not valid.
//
// rankers is a slice of ranks required to define the hand,
// it panics if the correct number of rankers is not supplied.
// If more than one rank is supplied for a ranker it uses the highest rank.
//
//   - A 6-quadruplet, 9-kicker four of a kind hand could be created with NewHand(FourOfAKind, Six, Nine).
//   - A 9-triple, 5-pair full house could be created with NewHand(FullHouse, Nine, Five).
//   - A Q-high straight could be created with NewHand(Straight, Queen).
func NewHand(hand Hand, rankers ...Cards) Hand {
	hand &= handKindMask
	numRankers := handRankers[hand>>28]
	if len(rankers) != numRankers {
		panic(
			fmt.Sprintf(
				"%s requires %v ranks to define, but %v was given",
				hand.Name(),
				rankers,
				len(rankers),
			),
		)
	}

	for i := range rankers {
		ri := len(rankers) - 1 - i
		hand |= Hand(rankers[i].RankAsInt()) << (ri * 4)
	}
	return hand
}

func new1ranker(hand Hand, ranker1 Cards) Hand {
	return hand |
		Hand(ranker1.RankAsInt())
}

func new2ranker(hand Hand, ranker1, ranker2 Cards) Hand {
	return hand |
		Hand(ranker1.RankAsInt())<<4 |
		Hand(ranker2.RankAsInt())
}

func new3ranker(hand Hand, ranker1, ranker2, ranker3 Cards) Hand {
	return hand |
		Hand(ranker1.RankAsInt())<<8 |
		Hand(ranker2.RankAsInt())<<4 |
		Hand(ranker3.RankAsInt())
}

func new4ranker(hand Hand, ranker1, ranker2, ranker3, ranker4 Cards) Hand {
	return hand |
		Hand(ranker1.RankAsInt())<<12 |
		Hand(ranker2.RankAsInt())<<8 |
		Hand(ranker3.RankAsInt())<<4 |
		Hand(ranker4.RankAsInt())
}

func new5ranker(hand Hand, ranker1, ranker2, ranker3, ranker4, ranker5 Cards) Hand {
	return hand |
		Hand(ranker1.RankAsInt())<<16 |
		Hand(ranker2.RankAsInt())<<12 |
		Hand(ranker3.RankAsInt())<<8 |
		Hand(ranker4.RankAsInt())<<4 |
		Hand(ranker5.RankAsInt())
}

// Rankers returns the ranks which define the hand.
// For each ranker it returns every possible
func (h Hand) Rankers() []Cards {
	rankers := make([]Cards, handRankers[h>>28])
	for i := range rankers {
		ri := len(rankers) - 1 - i
		rankers[i] = RankFromInt(int(h>>(ri*4)) & 0xF)
	}
	return rankers
}

// GetHand takes a set of cards and translates it into a hand.
//
// e.g. Ace is higher than 10, but four 10s obviously beats an Ace high card.
func GetHand(c Cards) Hand {
	// Straight Flush
	straight := Cards(0)
	straightFlushMask := AceSpades | KingSpades | QueenSpades | JackSpades | TenSpades
	for range 9 {
		if (c.Rank() & straightFlushMask).Len() == 5 {
			// we have a straight.
			// Check if it's a straight flush
			for s := 0; s < 4; s++ {
				if (c & (straightFlushMask >> s)).Len() == 5 {
					return new1ranker(StraightFlush, (c & (straightFlushMask >> s)).High())
				}
			}

			if straight == 0 {
				straight = (c & straightFlushMask.Rank()).High()
			}
		}

		straightFlushMask >>= 4
	}
	// steel wheel and baby straight
	if (c.Rank() & (AceSpades | FiveSpades | FourSpades | ThreeSpades | TwoSpades)).Len() == 5 {
		for s := 0; s < 4; s++ {
			if (c & ((AceSpades | FiveSpades | FourSpades | ThreeSpades | TwoSpades) >> s)).Len() == 5 {
				return new1ranker(StraightFlush, Five)
			}
		}

		if straight == 0 {
			straight = Five
		}
	}

	// get multiples
	//  - 0 unused
	//  - 1 highest card
	//  - 2 highest pair
	//  - 3 highest triple
	//  - 4 highest quadruple
	var multiples [5]Cards
	for rankInt := Two.RankAsInt(); rankInt <= Ace.RankAsInt(); rankInt++ {
		multiples[c.RankLen(rankInt)] = c & RankFromInt(rankInt)
	}

	// Four Of A Kind
	if multiples[4] != 0 {
		return new2ranker(FourOfAKind, multiples[4], (c &^ multiples[4]).High())
	}

	// Full House
	if multiples[3] != 0 && multiples[2] != 0 {
		return new2ranker(FullHouse, multiples[3], multiples[2])
	}

	// Flush
	for _, suite := range []Cards{Spades, Hearts, Diamonds, Clubs} {
		if (c & suite).Len() >= 5 {
			c &= suite
			card1 := c.High()
			card2 := c.NextHigh(card1)
			card3 := c.NextHigh(card2)
			card4 := c.NextHigh(card3)
			card5 := c.NextHigh(card4)
			return new5ranker(Flush, card1, card2, card3, card4, card5)
		}
	}

	// Straight
	if straight != 0 {
		return new1ranker(Straight, straight)
	}

	// Three Of A Kind
	if multiples[3] != 0 {
		card1 := (c &^ multiples[3]).High()
		card2 := (c &^ multiples[3]).NextHigh(card1)
		return new3ranker(ThreeOfAKind, multiples[3], card1, card2)
	}

	// Two Pair and One Pair
	if multiples[2] != 0 {
		pair := multiples[2]
		c &^= multiples[2]
		multiples[2] = 0

		for rankInt := Two.RankAsInt(); rankInt <= Ace.RankAsInt(); rankInt++ {
			multiples[c.RankLen(rankInt)] = c & RankFromInt(rankInt)
		}

		c &^= multiples[2]

		if multiples[2] != 0 {
			return new3ranker(TwoPair, pair, multiples[2], c.High())
		}

		card1 := c.High()
		card2 := c.NextHigh(card1)
		card3 := c.NextHigh(card2)
		return new4ranker(OnePair, pair, card1, card2, card3)
	}

	// Nothing
	card1 := c.High()
	card2 := c.NextHigh(card1)
	card3 := c.NextHigh(card2)
	card4 := c.NextHigh(card3)
	card5 := c.NextHigh(card4)
	return new5ranker(HighCard, card1, card2, card3, card4, card5)
}
