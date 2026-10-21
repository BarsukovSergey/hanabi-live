/* eslint-disable unicorn/no-null */

import { START_CARD_RANK } from "../constants";
import { StackDirection } from "../enums/StackDirection";
import type { CardState } from "../interfaces/CardState";
import type { GameState } from "../interfaces/GameState";
import type { Variant } from "../interfaces/Variant";
import type { Rank } from "../types/Rank";
import type { SuitIndex } from "../types/SuitIndex";
import { hasReversedSuits } from "./variants/variantIdentity";

/**
 * Returns an array since it is possible in some variants to have two or more possible cards that
 * are legal next plays.
 */
export function getNextPlayableRanks(
  suitIndex: SuitIndex,
  playStack: readonly number[],
  playStackDirection: StackDirection,
  playStackStarts: GameState["playStackStarts"],
  variant: Variant,
  deck: readonly CardState[],
): readonly number[] {
  const currentlyPlayedRank = getLastPlayedRank(playStack, deck);

  switch (playStackDirection) {
    case StackDirection.Undecided: {
      if (currentlyPlayedRank === START_CARD_RANK) {
        return [...new Set([2, variant.stackSize - 1])];
      }

      // With a three-rank stack, START -> 2 is still undecided:
      // playing either endpoint determines the direction and finishes the stack.
      if (currentlyPlayedRank !== null) {
        return [1, variant.stackSize];
      }

      return [...new Set([1, variant.stackSize, START_CARD_RANK])];
    }

    case StackDirection.Up: {
      if (!variant.sudoku) {
        // In non-Sudoku variants, the next playable card is 1 if the stack is not stared yet or the
        // N+1 rank.
        return currentlyPlayedRank === null ? [1] : [currentlyPlayedRank + 1];
      }

      // In Sudoku variants, determining the next playable ranks is more complicated. If the stack
      // is already started, then we go up, wrapping around from 5 to 1 (unless the stack was
      // started at 1, in which case 5 will be the last card of this suit).
      if (currentlyPlayedRank !== null) {
        // We mod by the stack size and then add to obtain values [1, ..., stackSize].
        return [(currentlyPlayedRank % variant.stackSize) + 1];
      }

      // The stack is not started yet. As a special case, we might already know the start of the
      // play stack, even when no cards have been played when this is the last suit. In that case,
      // only the (known) stack start can be played.
      const playStackStart = playStackStarts[suitIndex];
      if (playStackStart !== undefined && playStackStart !== null) {
        return [playStackStart];
      }

      // If the stack is not started, it can be started with any rank that is not the starting rank
      // of another stack.
      return variant.ranks.filter((rank) => !playStackStarts.includes(rank));
    }

    case StackDirection.Down: {
      // In non-Sudoku variants, start at the highest rank and play downward.
      return currentlyPlayedRank === null
        ? [variant.stackSize]
        : [currentlyPlayedRank - 1];
    }

    case StackDirection.Finished: {
      return [];
    }
  }
}

/** @returns `undefined` if there are no cards played on the stack. */
export function getLastPlayedRank(
  playStack: readonly number[],
  deck: readonly CardState[],
): Rank | null {
  const orderOfTopCard = playStack.at(-1);
  if (orderOfTopCard === undefined) {
    return null;
  }

  const card = deck[orderOfTopCard];
  if (card === undefined) {
    return null;
  }

  return card.rank;
}

export function getStackDirection(
  suitIndex: SuitIndex,
  playStack: readonly number[],
  deck: readonly CardState[],
  variant: Variant,
): StackDirection {
  if (playStack.length === variant.stackSize) {
    return StackDirection.Finished;
  }

  if (!hasReversedSuits(variant)) {
    return StackDirection.Up;
  }

  if (!variant.upOrDown) {
    const suit = variant.suits[suitIndex];
    if (suit === undefined) {
      return StackDirection.Up;
    }

    return suit.reversed ? StackDirection.Down : StackDirection.Up;
  }

  const firstRank = getStackStartRank(playStack, deck);
  if (firstRank === null) {
    return StackDirection.Undecided;
  }
  if (firstRank !== START_CARD_RANK) {
    // A regular Up or Down stack starts at either endpoint.
    return firstRank === 1 ? StackDirection.Up : StackDirection.Down;
  }

  // START replaces the endpoint. The following rank reveals the direction,
  // except for a three-rank stack, where START -> 2 can still go either way.
  const secondOrder = playStack[1];
  const secondRank =
    secondOrder === undefined ? undefined : deck[secondOrder]?.rank;
  if (secondRank === undefined) {
    return StackDirection.Undecided;
  }
  if (secondRank === 2 && secondRank === variant.stackSize - 1) {
    return StackDirection.Undecided;
  }

  return secondRank === 2 ? StackDirection.Up : StackDirection.Down;
}

/** Returns the rank of the bottom card of the stack. */
export function getStackStartRank(
  playStack: readonly number[],
  deck: readonly CardState[],
): Rank | null {
  const bottomCardOrder = playStack[0];
  if (bottomCardOrder === undefined) {
    return null;
  }

  const bottomCard = deck[bottomCardOrder];
  if (bottomCard === undefined) {
    return null;
  }

  return bottomCard.rank;
}
