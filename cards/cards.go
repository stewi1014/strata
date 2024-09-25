// cards provides methods for dealing with the [Standard 52-card deck]
//
// [Standard 52-card deck]: https://en.wikipedia.org/wiki/Standard_52-card_deck
package cards

//go:generate go run make_tables.go

import (
	"errors"
	"math/bits"
)

var (
	ErrInvalidSuite = errors.New("invalid suite")
	ErrInvalidRank  = errors.New("invalid rank")
)

const (
	Deck     = Cards(0xFFFFFFFFFFFFF) // Every card
	Spades   = Cards(0x8888888888888) // Every card of suite Spade
	Hearts   = Cards(0x4444444444444) // Every card of suite Heart
	Diamonds = Cards(0x2222222222222) // Every card of suite Diamond
	Clubs    = Cards(0x1111111111111) // Every card of suite Club
	Ace      = Cards(0xF000000000000) // Every card of rank Ace
	King     = Cards(0x0F00000000000) // Every card of rank King
	Queen    = Cards(0x00F0000000000) // Every card of rank Queen
	Jack     = Cards(0x000F000000000) // Every card of rank Jack
	Ten      = Cards(0x0000F00000000) // Every card of rank Ten
	Nine     = Cards(0x00000F0000000) // Every card of rank Nine
	Eight    = Cards(0x000000F000000) // Every card of rank Eight
	Seven    = Cards(0x0000000F00000) // Every card of rank Seven
	Six      = Cards(0x00000000F0000) // Every card of rank Six
	Five     = Cards(0x000000000F000) // Every card of rank Five
	Four     = Cards(0x0000000000F00) // Every card of rank Four
	Three    = Cards(0x00000000000F0) // Every card of rank Three
	Two      = Cards(0x000000000000F) // Every card of rank Two
)

const (
	TwoClubs      = Two & Clubs
	TwoDiamonds   = Two & Diamonds
	TwoHearts     = Two & Hearts
	TwoSpades     = Two & Spades
	ThreeClubs    = Three & Clubs
	ThreeDiamonds = Three & Diamonds
	ThreeHearts   = Three & Hearts
	ThreeSpades   = Three & Spades
	FourClubs     = Four & Clubs
	FourDiamonds  = Four & Diamonds
	FourHearts    = Four & Hearts
	FourSpades    = Four & Spades
	FiveClubs     = Five & Clubs
	FiveDiamonds  = Five & Diamonds
	FiveHearts    = Five & Hearts
	FiveSpades    = Five & Spades
	SixClubs      = Six & Clubs
	SixDiamonds   = Six & Diamonds
	SixHearts     = Six & Hearts
	SixSpades     = Six & Spades
	SevenClubs    = Seven & Clubs
	SevenDiamonds = Seven & Diamonds
	SevenHearts   = Seven & Hearts
	SevenSpades   = Seven & Spades
	EightClubs    = Eight & Clubs
	EightDiamonds = Eight & Diamonds
	EightHearts   = Eight & Hearts
	EightSpades   = Eight & Spades
	NineClubs     = Nine & Clubs
	NineDiamonds  = Nine & Diamonds
	NineHearts    = Nine & Hearts
	NineSpades    = Nine & Spades
	TenClubs      = Ten & Clubs
	TenDiamonds   = Ten & Diamonds
	TenHearts     = Ten & Hearts
	TenSpades     = Ten & Spades
	JackClubs     = Jack & Clubs
	JackDiamonds  = Jack & Diamonds
	JackHearts    = Jack & Hearts
	JackSpades    = Jack & Spades
	QueenClubs    = Queen & Clubs
	QueenDiamonds = Queen & Diamonds
	QueenHearts   = Queen & Hearts
	QueenSpades   = Queen & Spades
	KingClubs     = King & Clubs
	KingDiamonds  = King & Diamonds
	KingHearts    = King & Hearts
	KingSpades    = King & Spades
	AceClubs      = Ace & Clubs
	AceDiamonds   = Ace & Diamonds
	AceHearts     = Ace & Hearts
	AceSpades     = Ace & Spades
)

// Cards is a set of [Standard 52-card deck] cards. Like all [sets], it has no order and does not contain any duplicate elements.
//
// Cards are represented using bitflags, ordered based on the strength of the card.
// Cards can be compared accurately using integer comparason,
// and are designed to allow bitwise operations.
//
// The most significant 12 bits are ignored in arguments to methods,
// and are zeroed in return values unless otherwise specified.
//
// [Standard 52-card deck]: https://en.wikipedia.org/wiki/Standard_52-card_deck
// [sets]: https://en.wikipedia.org/wiki/Set_(mathematics)
type Cards uint64

// Len returns the number of cards in the set.
func (a Cards) Len() int {
	a &= Deck

	return bits.OnesCount64(uint64(a))
}

// High returns the highest card that exists in the set.
// If no cards exists it returns 0.
func (a Cards) High() Cards {
	a &= Deck

	return ((1 << 63) >> bits.LeadingZeros64(uint64(a))) & Deck
}

// NextHigh returns the highest card in a that is below all cards in b.
// If no cards exist it returns 0.
func (a Cards) NextHigh(b Cards) Cards {
	a &= Deck
	b &= Deck

	mask := ^Cards(0) >> (bits.LeadingZeros64(uint64(b)) + 1)
	return (a & mask).High()
}

// Low returns the lowest card that exists in the set.
// If no card exists it returns 0.
func (a Cards) Low() Cards {
	a &= Deck

	return (1 << bits.TrailingZeros64(uint64(a))) & Deck
}

// NextLow returns the lowest card in a that is above all cards in b.
func (a Cards) NextLow(b Cards) Cards {
	a &= Deck
	b &= Deck

	mask := ^Cards(0) >> bits.LeadingZeros64(uint64(b))
	return (a &^ mask).Low()
}

// Suite returns all cards that have the same suite.
func (a Cards) Suite() Cards {
	a &= Deck // Ignore 12MSBs

	a |= a >> 1 * 4
	a |= a >> 2 * 4
	a |= a >> 4 * 4
	a |= a >> 8 * 4

	a |= a << 1 * 4
	a |= a << 2 * 4
	a |= a << 4 * 4
	a |= a << 8 * 4
	return a & Deck
}

// Number returns the 52-card index of the highest card.
// It returns 0 if no card exists.
//
// E.g. TwoClubs = 1, TwoSpades = 4, ThreeDiamonds = 7
func (a Cards) Number() int {
	a &= Deck

	return bits.Len64(uint64(a))
}

// Rank returns all cards that have the same rank.
func (a Cards) Rank() Cards {
	a &= Deck

	r1 := a & 0xF0F0F0F0F0F0F0F0
	r1 |= r1 >> 1
	r1 |= r1 >> 2
	r1 |= r1 << 3
	r1 = r1 & 0xF0F0F0F0F0F0F0F0

	r2 := a & 0x0F0F0F0F0F0F0F0F
	r2 |= r2 << 1
	r2 |= r2 << 2
	r2 |= r2 >> 3
	r2 = r2 & 0x0F0F0F0F0F0F0F0F
	return Cards(r1 | r2)
}

// Rankers returns the cards with duplicate cards of the same rank removed.
// It returns the highest suite.
//
// E.g. (TwoSpades | TwoClubs | ThreeHearts).Rankers() = TwoSpades | ThreeHearts
func (a Cards) Rankers() Cards {
	b := Cards(0)
	b |= Cards(rankers8table[(a>>0)&0xFF]) << 0
	b |= Cards(rankers8table[(a>>8)&0xFF]) << 8
	b |= Cards(rankers8table[(a>>16)&0xFF]) << 16
	b |= Cards(rankers8table[(a>>24)&0xFF]) << 24
	b |= Cards(rankers8table[(a>>32)&0xFF]) << 32
	b |= Cards(rankers8table[(a>>40)&0xFF]) << 40
	b |= Cards(rankers8table[(a>>48)&0xFF]) << 48
	return b
}

// RankAsInt returns the highest rank as an integer.
// The rank is
//
//   - 14 = Ace
//   - 13 = King
//   - 12 = Queen
//   - 11 = Jack
//   - 10 = Ten
//   - 9  = Nine
//   - 8  = Eight
//   - 7  = Seven
//   - 6  = Six
//   - 5  = Five
//   - 4  = Four
//   - 3  = Three
//   - 2  = Two
//   - 1  = never returned
//   - 0  = no card exists
func (a Cards) RankAsInt() int {
	a &= Deck
	if a == 0 {
		return 0
	}

	return (bits.Len64(uint64(a)) + 7) / 4
}

// RankFromIntOld returns cards of the given rank.
// See RankAsInt for details on what ranks are.
func RankFromInt(rank int) Cards {
	if rank == 0 {
		return 0
	}

	if rank <= 1 || rank > 14 {
		panic(ErrInvalidRank)
	}

	return Two << ((rank - 2) * 4)
}

// RankLen returns the number of cards of the given rank.
// See RankAsInt for details on what ranks are.
func (a Cards) RankLen(rank int) int {
	return int(pop4table[(a>>((rank-2)*4))&0xF])
}

// Deal returns n random cards from the set.
//
// randIntN must return a non-negative number in the half-open interval [0,n),
// it is never called with n <= 0 or n >= 52
//
// If all cards are dealt it returns immediately.
func (a Cards) Deal(randIntN func(int) int, n int) (dealt Cards) {
	for range n {
		if a == 0 {
			return
		}

		n := randIntN(a.Len())
		card := a.High()
		for range n {
			card = a.NextHigh(card)
		}
		a &^= card
		dealt |= card
	}

	return
}

// SuiteAsInt returns the highest suite as an integer.
//
//   - 4 = Spades
//   - 3 = Hearts
//   - 2 = Diamonds
//   - 1 = Clubs
//   - 0 = no card exists
func (a Cards) SuiteAsInt() int {
	switch {
	case a&Spades > 0:
		return 4
	case a&Hearts > 0:
		return 3
	case a&Diamonds > 0:
		return 2
	case a&Clubs > 0:
		return 1
	default:
		return 0
	}
}

func SuiteFromInt(suite int) Cards {
	switch suite {
	case 4:
		return Spades
	case 3:
		return Hearts
	case 2:
		return Diamonds
	case 1:
		return Clubs
	case 0:
		return 0
	}

	panic(ErrInvalidSuite)
}
