package main

import "testing"

func TestVariableStackSizeReversiblePlay(t *testing.T) {
	tests := []struct {
		name           string
		stackSize      int
		upOrDown       bool
		reversed       bool
		ranks          []int
		wantDirections []int
	}{
		{"five-rank up after START", 5, true, false, []int{StartCardRank, 2, 3, 4, 5}, []int{StackDirectionUndecided, StackDirectionUp, StackDirectionUp, StackDirectionUp, StackDirectionFinished}},
		{"five-rank down after START", 5, true, false, []int{StartCardRank, 4, 3, 2, 1}, []int{StackDirectionUndecided, StackDirectionDown, StackDirectionDown, StackDirectionDown, StackDirectionFinished}},
		{"four-rank up", 4, true, false, []int{1, 2, 3, 4}, []int{StackDirectionUp, StackDirectionUp, StackDirectionUp, StackDirectionFinished}},
		{"four-rank down", 4, true, false, []int{4, 3, 2, 1}, []int{StackDirectionDown, StackDirectionDown, StackDirectionDown, StackDirectionFinished}},
		{"four-rank up after START", 4, true, false, []int{StartCardRank, 2, 3, 4}, []int{StackDirectionUndecided, StackDirectionUp, StackDirectionUp, StackDirectionFinished}},
		{"four-rank down after START", 4, true, false, []int{StartCardRank, 3, 2, 1}, []int{StackDirectionUndecided, StackDirectionDown, StackDirectionDown, StackDirectionFinished}},
		{"three-rank up after START", 3, true, false, []int{StartCardRank, 2, 3}, []int{StackDirectionUndecided, StackDirectionUndecided, StackDirectionFinished}},
		{"three-rank down after START", 3, true, false, []int{StartCardRank, 2, 1}, []int{StackDirectionUndecided, StackDirectionUndecided, StackDirectionFinished}},
		{"three-rank plain up", 3, true, false, []int{1, 2, 3}, []int{StackDirectionUp, StackDirectionUp, StackDirectionFinished}},
		{"three-rank plain down", 3, true, false, []int{3, 2, 1}, []int{StackDirectionDown, StackDirectionDown, StackDirectionFinished}},
		{"two-rank up after START", 2, true, false, []int{StartCardRank, 2}, []int{StackDirectionUndecided, StackDirectionFinished}},
		{"two-rank down after START", 2, true, false, []int{StartCardRank, 1}, []int{StackDirectionUndecided, StackDirectionFinished}},
		{"two-rank regular up", 2, true, false, []int{1, 2}, []int{StackDirectionUp, StackDirectionFinished}},
		{"two-rank regular down", 2, true, false, []int{2, 1}, []int{StackDirectionDown, StackDirectionFinished}},
		{"one-rank START", 1, true, false, []int{StartCardRank}, []int{StackDirectionFinished}},
		{"one-rank normal", 1, true, false, []int{1}, []int{StackDirectionFinished}},
		{"four-rank reversed", 4, false, true, []int{4, 3, 2, 1}, []int{StackDirectionDown, StackDirectionDown, StackDirectionDown, StackDirectionFinished}},
		{"three-rank reversed", 3, false, true, []int{3, 2, 1}, []int{StackDirectionDown, StackDirectionDown, StackDirectionFinished}},
		{"one-rank reversed", 1, false, true, []int{1}, []int{StackDirectionFinished}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			direction := StackDirectionUndecided
			if tc.reversed {
				direction = StackDirectionDown
			}
			g := &Game{
				Variant:             &Variant{StackSize: tc.stackSize, UpOrDown: tc.upOrDown, Suits: []*Suit{{Reversed: tc.reversed}}},
				Stacks:              []int{0},
				PlayStackDirections: []int{direction},
			}
			for i, rank := range tc.ranks {
				if variantReversiblePlay(g, NewCard(0, rank)) {
					t.Fatalf("step %d rank %d unexpectedly misplayed", i, rank)
				}
				g.Stacks[0] = rank
				if got, want := g.PlayStackDirections[0], tc.wantDirections[i]; got != want {
					t.Fatalf("step %d rank %d direction = %d, want %d", i, rank, got, want)
				}
			}
			if !variantReversiblePlay(g, NewCard(0, 1)) {
				t.Error("playing on a completed stack should fail")
			}
		})
	}
}

func TestThreeRankUpOrDownRemainsUndecidedAfterStartTwo(t *testing.T) {
	v := &Variant{StackSize: 3, UpOrDown: true, Suits: []*Suit{{}}}
	g := &Game{
		Variant:             v,
		Stacks:              []int{2},
		PlayStackDirections: []int{StackDirectionUndecided},
		Deck:                []*Card{NewCard(0, 3)},
	}
	if !variantReversiblePlay(g, NewCard(0, 2)) {
		t.Fatal("repeating rank 2 should be an invalid play")
	}
	if g.PlayStackDirections[0] != StackDirectionUndecided {
		t.Fatal("an invalid repeat must not lock the direction")
	}
	if variantReversibleCheckAllDead(g) {
		t.Fatal("a 3 remains playable after START -> 2")
	}
	g.Deck = []*Card{NewCard(0, 1)}
	if variantReversibleCheckAllDead(g) {
		t.Fatal("a 1 remains playable after START -> 2")
	}
	g.Deck = []*Card{NewCard(0, 2)}
	if !variantReversibleCheckAllDead(g) {
		t.Fatal("only a 2 remains; both directions are dead")
	}
}

func TestVariableStackSizeMaxScoreWalks(t *testing.T) {
	for _, size := range []int{2, 3, 4, 5} {
		g := &Game{Variant: &Variant{StackSize: size}}
		if got := variantReversibleWalkUp(g, map[int]bool{}); got != size {
			t.Errorf("size %d, upward max score = %d", size, got)
		}
		if got := variantReversibleWalkDown(g, map[int]bool{}); got != size {
			t.Errorf("size %d, downward max score = %d", size, got)
		}
		if got := variantReversibleWalkDown(g, map[int]bool{size: true}); got != 0 {
			t.Errorf("size %d, discarded starting rank: downward max score = %d", size, got)
		}
	}
}

func TestClueForCompletedVariableStack(t *testing.T) {
	tests := []struct {
		name      string
		variant   Variant
		rank      int
		direction int
		want      bool
	}{
		{"normal four finishes on four", Variant{StackSize: 4}, 4, StackDirectionUp, true},
		{"normal four does not finish on five", Variant{StackSize: 4}, 5, StackDirectionUp, false},
		{"reversed four finishes on one", Variant{StackSize: 4, Suits: []*Suit{{Reversed: true}}}, 1, StackDirectionFinished, true},
		{"reversed four finishes on four", Variant{StackSize: 4, Suits: []*Suit{{Reversed: true}}}, 4, StackDirectionFinished, true},
		{"reversed four not finished", Variant{StackSize: 4, Suits: []*Suit{{Reversed: true}}}, 4, StackDirectionDown, false},
		{"Up or Down three finishes on three", Variant{StackSize: 3, UpOrDown: true}, 3, StackDirectionFinished, true},
		{"Up or Down three finishes on one", Variant{StackSize: 3, UpOrDown: true}, 1, StackDirectionFinished, true},
		{"Up or Down one finishes with START", Variant{StackSize: 1, UpOrDown: true}, StartCardRank, StackDirectionFinished, true},
		{"Sudoku four finishes at any rank", Variant{StackSize: 4, Sudoku: true}, 2, StackDirectionFinished, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldAwardClueForStack(&tc.variant, tc.rank, tc.direction); got != tc.want {
				t.Errorf("clue = %v, want %v", got, tc.want)
			}
		})
	}
}
