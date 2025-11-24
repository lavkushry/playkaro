package teenpatti

import (
	"sort"
)

// Hand Types
const (
	Trail        = 6 // Three of a Kind (AAA > KKK > ... > 222)
	PureSequence = 5 // Straight Flush (AKQ > ... > 432 > 32A)
	Sequence     = 4 // Straight (AKQ > ... > 432 > 32A)
	Color        = 3 // Flush (A.. > K.. > ...)
	Pair         = 2 // Two of a Kind (AAK > ... > 223)
	HighCard     = 1 // High Card (A.. > K.. > ...)
)

// HandRank represents the strength of a hand
type HandRank struct {
	Type   int
	Values []int // Values to compare in case of tie
}

// EvaluateWinner determines the winner between two players
func EvaluateWinner(p1, p2 *Player) *Player {
	rank1 := GetHandRank(p1.Cards)
	rank2 := GetHandRank(p2.Cards)

	if rank1.Type > rank2.Type {
		return p1
	}
	if rank2.Type > rank1.Type {
		return p2
	}

	// Tie in type, compare values
	for i := 0; i < 3; i++ {
		if rank1.Values[i] > rank2.Values[i] {
			return p1
		}
		if rank2.Values[i] > rank1.Values[i] {
			return p2
		}
	}

	// Absolute tie (should be rare with single deck)
	return p1 // Default to challenger (or split pot in advanced logic)
}

// GetHandRank calculates the rank of a hand
func GetHandRank(cards []Card) HandRank {
	// Sort cards descending by value
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	v1, v2, v3 := sorted[0].Value, sorted[1].Value, sorted[2].Value
	s1, s2, s3 := sorted[0].Suit, sorted[1].Suit, sorted[2].Suit

	// 1. Trail (Three of a Kind)
	if v1 == v2 && v2 == v3 {
		return HandRank{Type: Trail, Values: []int{v1, v2, v3}}
	}

	// Check for Sequence
	isSeq := (v1 == v2+1 && v2 == v3+1) || (v1 == 14 && v2 == 3 && v3 == 2) // A-2-3 is valid sequence

	// 2. Pure Sequence (Straight Flush)
	if isSeq && s1 == s2 && s2 == s3 {
		return HandRank{Type: PureSequence, Values: []int{v1, v2, v3}}
	}

	// 3. Sequence (Straight)
	if isSeq {
		return HandRank{Type: Sequence, Values: []int{v1, v2, v3}}
	}

	// 4. Color (Flush)
	if s1 == s2 && s2 == s3 {
		return HandRank{Type: Color, Values: []int{v1, v2, v3}}
	}

	// 5. Pair
	if v1 == v2 {
		return HandRank{Type: Pair, Values: []int{v1, v3, v2}} // Pair value first, then kicker
	}
	if v2 == v3 {
		return HandRank{Type: Pair, Values: []int{v2, v1, v3}}
	}
	if v1 == v3 {
		return HandRank{Type: Pair, Values: []int{v1, v2, v3}}
	}

	// 6. High Card
	return HandRank{Type: HighCard, Values: []int{v1, v2, v3}}
}
