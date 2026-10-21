package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestRankModifierDeckCounts(t *testing.T) {
	tests := []struct {
		name         string
		numSuits     int
		rank1        int
		rank4        int
		totalDeck    int
		criticalRank int
		scarceOnes   bool
	}{
		{"Scarce Ones & Critical Fours (6 Suits)", 6, 2, 1, 48, 4, true},
		{"Scarce Ones (5 Suits)", 5, 2, 2, 45, 0, true},
		{"Critical Fours (5 Suits)", 5, 3, 1, 45, 4, false},
		{"No Variant", 5, 3, 2, 50, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			suits := make([]*Suit, tc.numSuits)
			for i := range suits {
				suits[i] = &Suit{}
			}
			variant := &Variant{
				Name:         tc.name,
				Suits:        suits,
				Ranks:        []int{1, 2, 3, 4, 5},
				StackSize:    5,
				CriticalRank: tc.criticalRank,
				ScarceOnes:   tc.scarceOnes,
			}
			game := &Game{Variant: variant, ExtraOptions: &ExtraOptions{}}
			game.InitDeck()

			if got := variant.GetDeckSize(); got != tc.totalDeck {
				t.Errorf("GetDeckSize() = %d, want %d", got, tc.totalDeck)
			}
			if got := len(game.Deck); got != tc.totalDeck {
				t.Errorf("InitDeck() created %d cards, want %d", got, tc.totalDeck)
			}
			for _, rank := range []int{1, 4} {
				want := tc.rank1
				if rank == 4 {
					want = tc.rank4
				}
				copies := 0
				for _, card := range game.Deck {
					if card.SuitIndex == 0 && card.Rank == rank {
						copies++
					}
				}
				if copies != want {
					t.Errorf("rank %d has %d copies in suit 0, want %d", rank, copies, want)
				}
			}
		})
	}
}

func TestCombinedRankModifierJSONFlags(t *testing.T) {
	var variant VariantJSON
	err := json.Unmarshal([]byte(`{"name":"Scarce Ones & Critical Fours (6 Suits)","scarceOnes":true,"criticalRank":4}`), &variant)
	if err != nil {
		t.Fatal(err)
	}
	if !variant.ScarceOnes || variant.CriticalRank != 4 {
		t.Fatalf("decoded modifiers: scarceOnes=%v, criticalRank=%d", variant.ScarceOnes, variant.CriticalRank)
	}
}

func TestCriticalRankForEveryRank(t *testing.T) {
	tests := []struct {
		criticalRank int
		scarceOnes   bool
		wantCounts   [5]int
	}{
		{1, false, [5]int{1, 2, 2, 2, 1}},
		{2, false, [5]int{3, 1, 2, 2, 1}},
		{3, false, [5]int{3, 2, 1, 2, 1}},
		{4, false, [5]int{3, 2, 2, 1, 1}},
		{5, false, [5]int{3, 2, 2, 2, 1}},
		{1, true, [5]int{1, 2, 2, 2, 1}},
		{4, true, [5]int{2, 2, 2, 1, 1}},
	}
	for _, tc := range tests {
		variant := &Variant{
			Name:         "Unrelated variant name",
			Suits:        []*Suit{{}},
			Ranks:        []int{1, 2, 3, 4, 5},
			StackSize:    5,
			CriticalRank: tc.criticalRank,
			ScarceOnes:   tc.scarceOnes,
		}
		wantTotal := 0
		for i, want := range tc.wantCounts {
			rank := i + 1
			wantTotal += want
			if got := numCopiesOfCard(variant.Suits[0], rank, variant); got != want {
				t.Errorf("criticalRank=%d scarceOnes=%v rank=%d: copies=%d, want=%d", tc.criticalRank, tc.scarceOnes, rank, got, want)
			}
		}
		if got := variant.GetDeckSize(); got != wantTotal {
			t.Errorf("criticalRank=%d scarceOnes=%v: GetDeckSize()=%d, want=%d", tc.criticalRank, tc.scarceOnes, got, wantTotal)
		}
		game := &Game{Variant: variant, ExtraOptions: &ExtraOptions{}}
		game.InitDeck()
		if len(game.Deck) != wantTotal {
			t.Errorf("criticalRank=%d scarceOnes=%v: InitDeck()=%d, want=%d", tc.criticalRank, tc.scarceOnes, len(game.Deck), wantTotal)
		}
	}
}

func TestSpecialDeckModes(t *testing.T) {
	tests := []struct {
		name       string
		ranks      []int
		suit       Suit
		upOrDown   bool
		sudoku     bool
		wantCopies []int
	}{
		{"Up or Down", []int{1, 2, 3, 4, 5, StartCardRank}, Suit{}, true, false, []int{1, 2, 2, 2, 1, 1}},
		{"Up or Down One of Each", []int{1, 2, 3, 4, 5, StartCardRank}, Suit{OneOfEach: true}, true, false, []int{1, 1, 1, 1, 1, 1}},
		{"Reversed", []int{1, 2, 3, 4, 5}, Suit{Reversed: true}, false, false, []int{1, 2, 2, 2, 3}},
		{"Sudoku", []int{1, 2, 3, 4}, Suit{}, false, true, []int{2, 2, 2, 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			variant := &Variant{
				Name:      "Unrelated variant name",
				Suits:     []*Suit{&tc.suit},
				Ranks:     tc.ranks,
				UpOrDown:  tc.upOrDown,
				Sudoku:    tc.sudoku,
				StackSize: 5,
			}
			wantTotal := 0
			for i, rank := range tc.ranks {
				got := numCopiesOfCard(variant.Suits[0], rank, variant)
				want := tc.wantCopies[i]
				if got != want {
					t.Errorf("rank %d: %d copies, want %d", rank, got, want)
				}
				wantTotal += want
			}
			game := &Game{Variant: variant, ExtraOptions: &ExtraOptions{}}
			game.InitDeck()
			if got := variant.GetDeckSize(); got != wantTotal {
				t.Errorf("GetDeckSize()=%d, want %d", got, wantTotal)
			}
			if len(game.Deck) != wantTotal {
				t.Errorf("InitDeck()=%d, want %d", len(game.Deck), wantTotal)
			}
		})
	}
}

func TestVariantPredicatesFollowFlags(t *testing.T) {
	tests := []struct {
		name   string
		set    func(*Variant)
		active func(*Variant) bool
	}{
		{"Alternating Clues", func(v *Variant) { v.AlternatingClues = true }, (*Variant).IsAlternatingClues},
		{"Clue Starved", func(v *Variant) { v.ClueStarved = true }, (*Variant).IsClueStarved},
		{"Cow & Pig", func(v *Variant) { v.CowAndPig = true }, (*Variant).IsCowAndPig},
		{"Duck", func(v *Variant) { v.Duck = true }, (*Variant).IsDuck},
		{"Throw It in a Hole", func(v *Variant) { v.ThrowItInAHole = true }, (*Variant).IsThrowItInAHole},
		{"Up or Down", func(v *Variant) { v.UpOrDown = true }, (*Variant).IsUpOrDown},
		{"Synesthesia", func(v *Variant) { v.Synesthesia = true }, (*Variant).IsSynesthesia},
		{"Critical Fours", func(v *Variant) { v.CriticalRank = 4 }, (*Variant).IsCriticalFours},
		{"Scarce Ones", func(v *Variant) { v.ScarceOnes = true }, (*Variant).IsScarceOnes},
		{"Sudoku", func(v *Variant) { v.Sudoku = true }, (*Variant).IsSudoku},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := &Variant{Name: tc.name + " (5 Suits)"}
			if tc.active(v) {
				t.Fatal("the name alone must not enable the variant behavior")
			}
			v.Name = "Unrelated variant name"
			tc.set(v)
			if !tc.active(v) {
				t.Fatal("the variant flag must enable behavior regardless of its name")
			}
		})
	}
	if got := (&Variant{Name: "Clue Starved (5 Suits)"}).GetAdjustedClueTokens(8); got != 8 {
		t.Errorf("name-only Clue Starved incorrectly adjusted tokens: %d", got)
	}
	if got := (&Variant{ClueStarved: true}).GetAdjustedClueTokens(8); got != 16 {
		t.Errorf("flagged Clue Starved adjusted tokens: %d, want 16", got)
	}
	if !(&Variant{Name: "Throw It in a Hole (5 Suits)"}).ShouldGiveClueTokenForFinishingStack() {
		t.Error("name-only hole variant incorrectly suppressed a clue token")
	}
	if (&Variant{ThrowItInAHole: true}).ShouldGiveClueTokenForFinishingStack() {
		t.Error("flagged hole variant should not grant a clue token")
	}
}

func TestCatalogCombinedVariantAndSpecialFlags(t *testing.T) {
	oldJSONPath, oldColors, oldSuits := jsonPath, colors, suits
	oldVariants, oldIDs, oldFullIDs := variants, variantIDMap, variantIDFullMap
	oldNames, oldOdd, oldEven := variantNames, oddClues, evenClues
	defer func() {
		jsonPath, colors, suits = oldJSONPath, oldColors, oldSuits
		variants, variantIDMap, variantIDFullMap = oldVariants, oldIDs, oldFullIDs
		variantNames, oddClues, evenClues = oldNames, oldOdd, oldEven
	}()

	jsonPath = filepath.Join("..", "..", "packages", "game", "src", "json")
	colorsInit()
	suitsInit()
	variantsInit()

	combined := variants["Scarce Ones & Critical Fours (6 Suits)"]
	if combined == nil {
		t.Fatal("missing the combined variant in variants.json")
	}
	if !combined.ScarceOnes || combined.CriticalRank != 4 {
		t.Fatalf("combined flags: scarceOnes=%v criticalRank=%d", combined.ScarceOnes, combined.CriticalRank)
	}
	if got := combined.GetDeckSize(); got != 48 {
		t.Errorf("combined deck size=%d, want 48", got)
	}
	game := &Game{Variant: combined, ExtraOptions: &ExtraOptions{}}
	game.InitDeck()
	if got := len(game.Deck); got != 48 {
		t.Errorf("combined initial deck size=%d, want 48", got)
	}

	clueStarved := variants["Clue Starved (6 Suits)"]
	if clueStarved == nil || !clueStarved.IsClueStarved() {
		t.Error("Clue Starved metadata was not loaded")
	}
	upOrDown := variants["Up or Down (5 Suits)"]
	if upOrDown == nil || !upOrDown.IsUpOrDown() || upOrDown.GetDeckSize() != 45 {
		t.Error("Up or Down metadata/rank counts were not loaded")
	}
	sudoku := variants["Sudoku (4 Suits)"]
	if sudoku == nil || !sudoku.IsSudoku() || sudoku.GetDeckSize() != 32 {
		t.Error("Sudoku metadata/rank counts were not loaded")
	}
}
