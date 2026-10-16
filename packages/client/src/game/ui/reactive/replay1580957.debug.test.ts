import { readFileSync } from "node:fs";
import type { GameAction, GameState } from "@hanabi-live/game";
import { test } from "@jest/globals";
import { loadGameJSON } from "../../../../test/loadGameJSON";
import { getSoundType } from "./getSoundType";

test("inspect exported replay turn 7", () => {
  const raw = JSON.parse(readFileSync("/tmp/hanabi-replay.json", "utf8")) as Record<string, unknown>;
  raw["options"] = { variant: "No Variant" };
  const state = loadGameJSON(raw as unknown as Parameters<typeof loadGameJSON>[0]);
  const states = state.replay.states;
  const a = { type: "play", order: 4, playerIndex: 0 } as GameAction;
  console.log("replay states", states.length);
  for (let i = 5; i <= 9; i++) {
    const before = states[i - 1] as GameState | undefined;
    const after = states[i] as GameState | undefined;
    if (before === undefined || after === undefined) continue;
    console.log("segment", i, "turn",before.turn.turnNum,after.turn.turnNum,
      "stacks", after.playStacks.map(x=>x.length),
      "hand",before.hands[0],
      "right",before.deck[0] === undefined ? null : {
        positiveRankClues:before.deck[0].positiveRankClues,
        positiveColorClues:before.deck[0].positiveColorClues,
        dealtToStartingHand:before.deck[0].dealtToStartingHand,
        numPositiveClues:before.deck[0].numPositiveClues
      },
      "left",before.deck[4] === undefined ? null : {
        positiveRankClues:before.deck[4].positiveRankClues,
        positiveColorClues:before.deck[4].positiveColorClues,
        dealtToStartingHand:before.deck[4].dealtToStartingHand,
        numPositiveClues:before.deck[4].numPositiveClues
      },
      "sound",getSoundType(before,after,a,state.metadata));
  }
});
