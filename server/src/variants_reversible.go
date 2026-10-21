// Helper functions for variants where suits have a non-standard playing direction
// (e.g. 5 --> 4 --> 3 --> 2 --> 1)
// Currently used for "Up Or Down" and "Reversed" variants

package main

const (
	// The "Up or Down" variants have "START" cards
	// Rank 0 is the stack base
	// Rank 1-5 are the normal cards
	// Rank 6 is a card of unknown rank
	// Rank 7 is a "START" card
	StartCardRank = 7
)

func variantReversiblePlay(g *Game, c *Card) bool {
	variant := g.Variant
	finalRank := variant.StackSize
	lastRank := g.Stacks[c.SuitIndex]
	direction := &g.PlayStackDirections[c.SuitIndex]

	switch *direction {
	case StackDirectionUndecided:
		if lastRank == 0 {
			if c.Rank != 1 && c.Rank != finalRank && c.Rank != StartCardRank {
				return true
			}
			switch {
			case finalRank == 1:
				*direction = StackDirectionFinished
			case c.Rank == 1:
				*direction = StackDirectionUp
			case c.Rank == finalRank:
				*direction = StackDirectionDown
			}
			return false
		}

		if lastRank == StartCardRank {
			if c.Rank != 2 && c.Rank != finalRank-1 {
				return true
			}
			switch {
			case finalRank == 2:
				// START plus either endpoint completes the two-rank stack.
				*direction = StackDirectionFinished
			case c.Rank == 2 && c.Rank == finalRank-1:
				// For three ranks, START -> 2 can still go up or down.
			case c.Rank == 2:
				*direction = StackDirectionUp
			default:
				*direction = StackDirectionDown
			}
			return false
		}

		if finalRank == 3 && lastRank == 2 {
			// Complete either direction after the ambiguous START -> 2.
			if c.Rank == 1 || c.Rank == finalRank {
				*direction = StackDirectionFinished
				return false
			}
		}
		return true

	case StackDirectionUp:
		if c.Rank != lastRank+1 {
			return true
		}
		if c.Rank == finalRank {
			*direction = StackDirectionFinished
		}
		return false

	case StackDirectionDown:
		nextRank := lastRank - 1
		if !variant.IsUpOrDown() && lastRank == 0 {
			nextRank = finalRank
		}
		if c.Rank != nextRank {
			return true
		}
		if c.Rank == 1 {
			*direction = StackDirectionFinished
		}
		return false
	}

	return true
}

// variantReversibleGetMaxScore calculates what the maximum score is,
// accounting for stacks that cannot be completed due to discarded cards
func variantReversibleGetMaxScore(g *Game) int {
	// Local variables
	variant := g.Variant

	maxScore := 0
	for suitIndex := range g.Stacks {
		// Check every rank available in the current variant, including START.
		allDiscarded := make(map[int]bool)
		for _, rank := range variant.Ranks {
			total, discarded := g.GetSpecificCardNum(suitIndex, rank)
			allDiscarded[rank] = total == discarded
		}

		if g.PlayStackDirections[suitIndex] == StackDirectionUndecided {
			upWalk := variantReversibleWalkUp(g, allDiscarded)
			downWalk := variantReversibleWalkDown(g, allDiscarded)
			maxScore += max(upWalk, downWalk)
		} else if g.PlayStackDirections[suitIndex] == StackDirectionUp {
			maxScore += variantReversibleWalkUp(g, allDiscarded)
		} else if g.PlayStackDirections[suitIndex] == StackDirectionDown {
			maxScore += variantReversibleWalkDown(g, allDiscarded)
		} else if g.PlayStackDirections[suitIndex] == StackDirectionFinished {
			maxScore += variant.StackSize
		}
	}

	return maxScore
}

// A helper function for "variantReversibleGetMaxScore()"
func variantReversibleWalkUp(g *Game, allDiscarded map[int]bool) int {
	// Local variables
	variant := g.Variant

	cardsThatCanStillBePlayed := 0

	// First, check to see if the stack can still be started
	if variant.IsUpOrDown() {
		if allDiscarded[1] && allDiscarded[StartCardRank] {
			// In "Up or Down" variants, you can start with 1 or START when going up
			return 0
		}
	} else {
		if allDiscarded[1] {
			// Otherwise, only 1
			return 0
		}
	}
	cardsThatCanStillBePlayed++

	// Second, walk upwards
	for rank := 2; rank <= variant.StackSize; rank++ {
		if allDiscarded[rank] {
			break
		}
		cardsThatCanStillBePlayed++
	}

	return cardsThatCanStillBePlayed
}

// A helper function for "variantReversibleGetMaxScore()"
func variantReversibleWalkDown(g *Game, allDiscarded map[int]bool) int {
	// Local variables
	variant := g.Variant

	cardsThatCanStillBePlayed := 0

	// First, check to see if the stack can still be started
	if variant.IsUpOrDown() {
		if allDiscarded[variant.StackSize] && allDiscarded[StartCardRank] {
			// In "Up or Down" variants, you can start with 5 or START when going down
			return 0
		}
	} else {
		if allDiscarded[variant.StackSize] {
			// Otherwise, only 5
			return 0
		}
	}
	cardsThatCanStillBePlayed++

	// Second, walk downwards
	for rank := variant.StackSize - 1; rank >= 1; rank-- {
		if allDiscarded[rank] {
			break
		}
		cardsThatCanStillBePlayed++
	}

	return cardsThatCanStillBePlayed
}

// variantReversibleCheckAllDead returns true if no more cards can be played on the stacks
func variantReversibleCheckAllDead(g *Game) bool {
	// Local variables
	variant := g.Variant

	for suitIndex, stackRank := range g.Stacks {
		neededRanks := make([]int, 0)
		if g.PlayStackDirections[suitIndex] == StackDirectionUndecided {
			if stackRank == 0 {
				// Nothing is played on the stack
				neededRanks = []int{1, variant.StackSize, StartCardRank}
			} else if stackRank == StartCardRank {
				// The START card is played; either direction may be chosen.
				neededRanks = []int{2, variant.StackSize - 1}
			} else if variant.StackSize == 3 && stackRank == 2 {
				// START -> 2 leaves both directions available on three ranks.
				neededRanks = []int{1, variant.StackSize}
			}
		} else if g.PlayStackDirections[suitIndex] == StackDirectionUp {
			neededRanks = append(neededRanks, stackRank+1)
		} else if g.PlayStackDirections[suitIndex] == StackDirectionDown {
			if !variant.IsUpOrDown() && stackRank == 0 {
				// Reversed stacks start at the highest rank.
				neededRanks = []int{variant.StackSize}
			} else {
				neededRanks = append(neededRanks, stackRank-1)
			}
		} else if g.PlayStackDirections[suitIndex] == StackDirectionFinished {
			continue
		}

		for _, c := range g.Deck {
			for _, neededRank := range neededRanks {
				if c.SuitIndex == suitIndex &&
					c.Rank == neededRank &&
					!c.Discarded &&
					!c.CannotBePlayed {

					return false
				}
			}
		}
	}

	// If we got this far, nothing can be played
	return true
}
