package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// This is an independent reference for the copy-count rules used by the
// TypeScript getNumCopiesOfCard implementation in packages/game/src/rules/deck.ts.
func expectedClientCardCopies(meta VariantJSON, suit *Suit, rank int) int {
	if suit.OneOfEach || (meta.CriticalRank != 0 && meta.CriticalRank == rank) {
		return 1
	}
	if meta.Sudoku {
		return 2
	}
	if rank == StartCardRank {
		if meta.UpOrDown {
			return 1
		}
		return -1
	}
	if rank == 1 {
		if meta.UpOrDown || suit.Reversed {
			return 1
		}
		if meta.ScarceOnes {
			return 2
		}
		return 3
	}
	stackSize := meta.StackSize
	if stackSize == 0 {
		stackSize = DefaultPointsPerSuit
	}
	if rank == stackSize {
		if suit.Reversed {
			return 3
		}
		return 1
	}
	return 2
}

// These two reference functions follow packages/game/src/rules/clues.ts.
// The expected behavior comes from JSON metadata rather than server predicates.
func expectedClientColorTouch(meta VariantJSON, variant *Variant, suit *Suit, cardRank int, clueColorName string) bool {
	if meta.ColorCluesTouchNothing {
		return false
	}
	if suit.AllClueColors {
		return true
	}
	if suit.NoClueColors {
		return false
	}
	if meta.Synesthesia && !suit.NoClueRanks && len(variant.ClueColors) > 0 {
		colorIndex := (cardRank - 1) % len(variant.ClueColors)
		if variant.ClueColors[colorIndex] == clueColorName {
			return true
		}
	}
	if cardRank == meta.SpecialRank {
		if meta.SpecialRankAllClueColors {
			return true
		}
		if meta.SpecialRankNoClueColors {
			return false
		}
	}
	if suit.Prism && len(variant.ClueColors) > 0 {
		colorIndex := (cardRank - 1) % len(variant.ClueColors)
		if cardRank == StartCardRank {
			colorIndex = len(variant.ClueColors) - 1
		}
		return variant.ClueColors[colorIndex] == clueColorName
	}
	return stringInSlice(clueColorName, suit.ClueColors)
}

func expectedClientRankTouch(meta VariantJSON, variant *Variant, suitIndex int, suit *Suit, cardRank int, clueRank int) bool {
	if meta.RankCluesTouchNothing {
		return false
	}
	if suit.AllClueRanks {
		return true
	}
	if suit.NoClueRanks {
		return false
	}
	if meta.Funnels {
		return cardRank <= clueRank
	}
	if meta.Chimneys {
		return cardRank >= clueRank
	}
	if meta.OddsAndEvens {
		if clueRank == 1 {
			return cardRank == 1 || cardRank == 3 || cardRank == 5
		}
		return cardRank == 2 || cardRank == 4
	}
	if cardRank == meta.SpecialRank {
		if meta.SpecialRankAllClueRanks {
			return true
		}
		if meta.SpecialRankNoClueRanks {
			return false
		}
		if meta.SpecialRankDeceptive {
			if len(variant.ClueRanks) == 0 {
				return false
			}
			return clueRank == variant.ClueRanks[suitIndex%len(variant.ClueRanks)]
		}
	}
	return clueRank == cardRank
}

func TestCatalogMatchesClientVariantMetadata(t *testing.T) {
	oldPath, oldColors, oldSuits := jsonPath, colors, suits
	oldVariants, oldIDMap, oldIDFullMap := variants, variantIDMap, variantIDFullMap
	oldNames, oldOddClues, oldEvenClues := variantNames, oddClues, evenClues
	defer func() {
		jsonPath, colors, suits = oldPath, oldColors, oldSuits
		variants, variantIDMap, variantIDFullMap = oldVariants, oldIDMap, oldIDFullMap
		variantNames, oddClues, evenClues = oldNames, oldOddClues, oldEvenClues
	}()

	jsonPath = filepath.Join("..", "..", "packages", "game", "src", "json")
	colorsInit()
	suitsInit()
	variantsInit()

	data, err := os.ReadFile(filepath.Join(jsonPath, "variants.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog []VariantJSON
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatal(err)
	}
	if len(variants) != len(catalog) {
		t.Fatalf("loaded %d variants, expected %d catalog entries", len(variants), len(catalog))
	}

	for _, meta := range catalog {
		v := variants[meta.Name]
		if v == nil {
			t.Errorf("variant %q was not loaded", meta.Name)
			continue
		}

		// All special-rule predicates must come from catalog metadata, not names.
		modes := []struct {
			name string
			got  bool
			want bool
		}{
			{"colorCluesTouchNothing", v.ColorCluesTouchNothing, meta.ColorCluesTouchNothing},
			{"rankCluesTouchNothing", v.RankCluesTouchNothing, meta.RankCluesTouchNothing},
			{"specialRankAllClueColors", v.SpecialRankAllClueColors, meta.SpecialRankAllClueColors},
			{"specialRankAllClueRanks", v.SpecialRankAllClueRanks, meta.SpecialRankAllClueRanks},
			{"specialRankNoClueColors", v.SpecialRankNoClueColors, meta.SpecialRankNoClueColors},
			{"specialRankNoClueRanks", v.SpecialRankNoClueRanks, meta.SpecialRankNoClueRanks},
			{"specialRankDeceptive", v.SpecialRankDeceptive, meta.SpecialRankDeceptive},
			{"alternatingClues", v.IsAlternatingClues(), meta.AlternatingClues},
			{"clueStarved", v.IsClueStarved(), meta.ClueStarved},
			{"cowAndPig", v.IsCowAndPig(), meta.CowAndPig},
			{"duck", v.IsDuck(), meta.Duck},
			{"throwItInAHole", v.IsThrowItInAHole(), meta.ThrowItInAHole},
			{"upOrDown", v.IsUpOrDown(), meta.UpOrDown},
			{"synesthesia", v.IsSynesthesia(), meta.Synesthesia},
			{"sudoku", v.IsSudoku(), meta.Sudoku},
			{"scarceOnes", v.IsScarceOnes(), meta.ScarceOnes},
			{"criticalFours", v.IsCriticalFours(), meta.CriticalRank == 4},
			{"funnels", v.Funnels, meta.Funnels},
			{"chimneys", v.Chimneys, meta.Chimneys},
			{"oddsAndEvens", v.OddsAndEvens, meta.OddsAndEvens},
		}
		for _, mode := range modes {
			if mode.got != mode.want {
				t.Errorf("%s: %s = %v; expected %v from variants.json", meta.Name, mode.name, mode.got, mode.want)
			}
		}
		if v.CriticalRank != meta.CriticalRank {
			t.Errorf("%s: critical rank = %d, want %d", meta.Name, v.CriticalRank, meta.CriticalRank)
		}
		wantSpecialRank := meta.SpecialRank
		if wantSpecialRank == 0 {
			wantSpecialRank = -1
		}
		if v.SpecialRank != wantSpecialRank {
			t.Errorf("%s: special rank = %d, want %d", meta.Name, v.SpecialRank, wantSpecialRank)
		}
		wantClueRanks := meta.ClueRanks
		if wantClueRanks == nil {
			defaults := []int{1, 2, 3, 4, 5}
			wantClueRanks = &defaults
		}
		if !reflect.DeepEqual(v.ClueRanks, *wantClueRanks) {
			t.Errorf("%s: clue ranks = %v, want %v", meta.Name, v.ClueRanks, *wantClueRanks)
		}

		stackSize := meta.StackSize
		if stackSize == 0 {
			stackSize = DefaultPointsPerSuit
		}
		if v.StackSize != stackSize {
			t.Errorf("%s: stack size = %d, want %d", meta.Name, v.StackSize, stackSize)
		}
		wantRanks := make([]int, stackSize)
		for i := range wantRanks {
			wantRanks[i] = i + 1
		}
		if meta.UpOrDown {
			wantRanks = append(wantRanks, StartCardRank)
		}
		if !reflect.DeepEqual(v.Ranks, wantRanks) {
			t.Errorf("%s: ranks = %v, want %v", meta.Name, v.Ranks, wantRanks)
		}

		// Validate clue-touch behavior across every available suit, rank, and clue.
		// This also checks the rank-modifier precedence for actual catalog variants.
		for suitIndex, suit := range v.Suits {
			for _, rank := range v.Ranks {
				card := NewCard(suitIndex, rank)
				for colorIndex, colorName := range v.ClueColors {
					clue := Clue{Type: ClueTypeColor, Value: colorIndex}
					wantTouch := expectedClientColorTouch(meta, v, suit, rank, colorName)
					if got := variantIsCardTouched(v, clue, card); got != wantTouch {
						t.Errorf("%s: suit %s rank %d color %s touches=%v, want %v", meta.Name, suit.Name, rank, colorName, got, wantTouch)
					}
				}
				for _, clueRank := range v.ClueRanks {
					clue := Clue{Type: ClueTypeRank, Value: clueRank}
					wantTouch := expectedClientRankTouch(meta, v, suitIndex, suit, rank, clueRank)
					if got := variantIsCardTouched(v, clue, card); got != wantTouch {
						t.Errorf("%s: suit %s rank %d clue rank %d touches=%v, want %v", meta.Name, suit.Name, rank, clueRank, got, wantTouch)
					}
				}
			}
		}

		wantDeckSize := 0
		for _, suit := range v.Suits {
			for _, rank := range wantRanks {
				wantCopies := expectedClientCardCopies(meta, suit, rank)
				if wantCopies < 0 {
					t.Errorf("%s: illegal START rank in non-Up-or-Down variant", meta.Name)
					continue
				}
				wantDeckSize += wantCopies
				if got := numCopiesOfCard(suit, rank, v); got != wantCopies {
					t.Errorf("%s: suit %s rank %d has %d copies, want %d", meta.Name, suit.Name, rank, got, wantCopies)
				}
			}
		}
		if got := v.GetDeckSize(); got != wantDeckSize {
			t.Errorf("%s: deck size = %d, want %d", meta.Name, got, wantDeckSize)
		}
		game := &Game{Variant: v, ExtraOptions: &ExtraOptions{}}
		game.InitDeck()
		if len(game.Deck) != wantDeckSize || len(game.CardIdentities) != wantDeckSize {
			t.Errorf("%s: created %d cards and %d identities, want %d", meta.Name, len(game.Deck), len(game.CardIdentities), wantDeckSize)
		}
	}
}
