import { SoundMove } from "@hanabi-live/data";
import type { GameAction, GameState } from "@hanabi-live/game";
import { beforeEach, describe, expect, jest, test } from "@jest/globals";
import { SoundType } from "../../../types/SoundType";
import { globals } from "../../UIGlobals";
import { onNewSoundEffect } from "./soundView";

jest.mock("../../UIGlobals", () => ({
  globals: {
    isResizing: false,
    lobby: {
      settings: { soundMove: SoundMove.All, hyphenatedConventions: true },
    },
    variant: { throwItInAHole: false },
    state: {
      replay: { hypothetical: null, states: [] },
      finished: false,
    },
    metadata: {
      options: { variantName: "No Variant" },
      ourPlayerIndex: 0,
    },
    game: { sounds: { play: jest.fn() } },
  },
}));
jest.mock("../../isOurTurn", () => ({ isOurTurn: () => true }));

describe("OCM sound playback", () => {
  const olderOne = {
    order: 0,
    numPositiveClues: 1,
    positiveRankClues: [1],
    positiveColorClues: [],
    dealtToStartingHand: true,
  };
  const newerOne = { ...olderOne, order: 4 };

  const previousState = {
    deck: [olderOne, undefined, undefined, undefined, newerOne],
    hands: [[4, 0]],
    stats: { maxScore: 25 },
    turn: { currentPlayerIndex: 0, turnNum: 7 },
  } as unknown as GameState;
  const currentState = {
    deck: previousState.deck,
    hands: [[0]],
    stats: { maxScore: 25, numSubsequentBlindPlays: 0 },
    playStacks: [[1], [1], [1], [1], []],
    turn: { currentPlayerIndex: 1, turnNum: 8 },
  } as unknown as GameState;
  const action = {
    type: "play" as const,
    order: 4,
    playerIndex: 0,
    suitIndex: 3,
    rank: 1,
  } as unknown as GameAction;

  beforeEach(() => {
    jest.clearAllMocks();
    Object.assign(globals.state.replay, {
      states: [previousState, currentState],
    });
    Object.assign(globals.lobby.settings, { soundMove: SoundMove.All });
  });

  test.each([
    [true, SoundType.OrderChopMove],
    [false, SoundType.Other],
  ])("convention setting %s yields sound %s", (hyphenated, expected) => {
    Object.assign(globals.lobby.settings, {
      hyphenatedConventions: hyphenated,
    });

    onNewSoundEffect(
      { gameState: currentState, actions: [action] },
      { gameState: previousState, actions: [] },
    );

    expect(globals.game!.sounds.play).toHaveBeenCalledWith(expected, false);
  });
});
