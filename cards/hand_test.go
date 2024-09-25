package cards

import (
	"testing"
)

var handSink Hand

var testHands = []struct {
	cards string
	hand  string
}{
	{
		cards: "KC QS 7D 5H 4S 3H 2C",
		hand:  "High Card (K, Q, 7, 5, 4)",
	},
	{
		cards: "AC KC QS JH 4S 3H 2C",
		hand:  "High Card (A, K, Q, J, 4)",
	},
	{
		cards: "AC KC QS JH 9S 3H 2C",
		hand:  "High Card (A, K, Q, J, 9)",
	},
	{
		cards: "8h 9d 4h 4c 7c qc ah",
		hand:  "One Pair (4, A, Q, 9)",
	},
	{
		cards: "5H 6S 2D 3C 9H AC 6D",
		hand:  "One Pair (6, A, 9, 5)",
	},
	{
		cards: "KH QC JS JC TS TC 6D",
		hand:  "Two Pair (J, T, K)",
	},
	{
		cards: "QS QC JC TC 9S 9H 3H",
		hand:  "Two Pair (Q, 9, J)",
	},
	{
		cards: "AH QS QH QC JC TC 5S",
		hand:  "Three Of A Kind (Q, A, J)",
	},
	{
		cards: "AC TC 5S 5H 4S 3H 2C",
		hand:  "Straight (5)",
	},
	{
		cards: "AC KC QS JH TS 3H 3C",
		hand:  "Straight (A)",
	},
	{
		cards: "AH KD QC JC TC 8H 2C",
		hand:  "Straight (A)",
	},
	{
		cards: "AH KD QH TC 6H 5H 2H",
		hand:  "Flush (A, Q, 6, 5, 2)",
	},
	{
		cards: "AH QH TC 6H 5H 3H 2H",
		hand:  "Flush (A, Q, 6, 5, 3)",
	},
	{
		cards: "AH QH TC 6H 5H 3H 2C",
		hand:  "Flush (A, Q, 6, 5, 3)",
	},
	{
		cards: "KH QC JS JC TS TC TD",
		hand:  "Full House (T, J)",
	},
	{
		cards: "AC AH QS QH QC TC 5S",
		hand:  "Full House (Q, A)",
	},
	{
		cards: "AC AH AS QH QC TC 5S",
		hand:  "Full House (A, Q)",
	},
	{
		cards: "QS QH QD QC JC TC 5S",
		hand:  "Four Of A Kind (Q, J)",
	},
	{
		cards: "AC TC 5C 5H 4C 3C 2C",
		hand:  "Straight Flush (5)",
	},
	{
		cards: "AS TC 5C 5S 4S 3S 2S",
		hand:  "Straight Flush (5)",
	},
	{
		cards: "AC KC QC JC TC 9H 2C",
		hand:  "Royal Flush",
	},
}

func BenchmarkGetHand(b *testing.B) {
	testCards := make([]Cards, len(testHands))
	for i := range testHands {
		testCards[i] = ParseCards(testHands[i].cards)
	}

	for i := 0; i < b.N; i++ {
		for _, cards := range testCards {
			handSink = GetHand(cards)
		}
	}
}

func TestGetHand(t *testing.T) {
	parsedHands := make([]Hand, len(testHands))

	for i, tt := range testHands {
		t.Run(tt.hand, func(t *testing.T) {
			cards := ParseCards(tt.cards)
			got := GetHand(cards)

			if got.String() != tt.hand {
				t.Errorf("GetHand() = %v, want %v", got.String(), tt.hand)
			}

			parsedHands[i] = got
		})
	}

	lastHand := parsedHands[0]
	for _, hand := range parsedHands {
		if hand < lastHand {
			// hands in testHands should be ordered from weakest to strongest.
			t.Errorf(
				"Hand %v is smaller than %v, but should be larger",
				hand,
				lastHand,
			)
		} else {
			lastHand = hand
		}
	}
}
