import type {
  ActionPlay,
  CardState,
  GameMetadata,
  GameState,
} from "@hanabi-live/game";
import { describe, expect, test } from "@jest/globals";
import { SoundType } from "../../types/SoundType";
import { getSoundType } from "./getSoundType";

function makeOne(order: number, suitIndex: number): CardState {
  return {
    order,
    suitIndex,
    rank: 1,
    numPositiveClues: 1,
    positiveRankClues: [1],
    positiveColorClues: [],
    dealtToStartingHand: true,
  } as unknown as CardState;
}

describe("OCM sound, replay 1580957 turn 7", () => {
  const playedBlueOne = makeOne(4, 3);
  const skippedRedOne = makeOne(0, 0);
  const metadata = {
    options: { variantName: "No Variant" },
    ourPlayerIndex: 0,
  } as GameMetadata;

  test.each([true, false])(
    "plays OCM when a newer clued starting 1 skips an older 1 (trash: %s)",
    (skippedIsTrash) => {
      const previous = {
        deck: [skippedRedOne, undefined, undefined, undefined, playedBlueOne],
        hands: [[4, 0]],
        stats: { maxScore: 25 },
      } as unknown as GameState;
      const current = {
        deck: previous.deck,
        turn: { currentPlayerIndex: 1 },
        stats: { maxScore: 25, numSubsequentBlindPlays: 0 },
        // After Blue 1, only Green 1 remains unplayed in replay 1580957.
        playStacks: [skippedIsTrash ? [1] : [], [1], [], [1], [1]],
      } as unknown as GameState;
      const action = { type: "play", order: 4, playerIndex: 0 } as ActionPlay;

      expect(getSoundType(previous, current, action, metadata)).toBe(
        SoundType.OrderChopMove,
      );
    },
  );
});
