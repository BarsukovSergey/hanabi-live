import type {
  ActionPlay,
  CardState,
  GameMetadata,
  GameState,
} from "@hanabi-live/game";
import { describe, expect, test } from "@jest/globals";
import { SoundType } from "../../types/SoundType";
import { getSoundType } from "./getSoundType";

describe("OCM sound", () => {
  const metadata = {
    options: { variantName: "No Variant" },
    ourPlayerIndex: 0,
  } as GameMetadata;

  const olderOne = {
    order: 0,
    positiveRankClues: [1],
    positiveColorClues: [],
    numPositiveClues: 1,
    dealtToStartingHand: true,
  } as unknown as CardState;

  const newerOne = {
    order: 1,
    positiveRankClues: [1],
    positiveColorClues: [],
    numPositiveClues: 1,
    dealtToStartingHand: true,
  } as unknown as CardState;

  const action = {
    type: "play",
    order: 1,
    playerIndex: 0,
  } as ActionPlay;

  test.each([
    ["only one playable 1 remains", [1, 1, 1, 1, 0]],
    ["multiple playable 1s remain", [1, 1, 0, 0, 0]],
  ])("plays OCM when a newer clued 1 is played and %s", (_name, lengths) => {
    const previous = {
      deck: [olderOne, newerOne],
      hands: [[0, 1]],
      stats: { maxScore: 25 },
    } as unknown as GameState;
    const current = {
      deck: [olderOne, newerOne],
      hands: [[0]],
      stats: { maxScore: 25 },
      turn: { currentPlayerIndex: 1 },
      playStacks: lengths.map((length) => Array.from({ length })),
    } as unknown as GameState;

    expect(getSoundType(previous, current, action, metadata)).toBe(
      SoundType.OrderChopMove,
    );
  });

  test("does not play OCM for the oldest clued 1", () => {
    const previous = {
      deck: [olderOne, newerOne],
      hands: [[0, 1]],
      stats: { maxScore: 25 },
    } as unknown as GameState;
    const current = {
      deck: [olderOne, newerOne],
      stats: { maxScore: 25 },
      turn: { currentPlayerIndex: 1 },
      playStacks: [[], [], [], [], []],
    } as unknown as GameState;
    const oldestPlay = { ...action, order: 0 } as ActionPlay;

    expect(getSoundType(previous, current, oldestPlay, metadata)).toBe(
      SoundType.Other,
    );
  });
});
