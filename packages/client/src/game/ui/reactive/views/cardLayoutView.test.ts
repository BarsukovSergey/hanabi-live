import type { CardOrder } from "@hanabi-live/game";
import { beforeEach, describe, expect, jest, test } from "@jest/globals";
import type { HanabiCard } from "../../HanabiCard";
import { globals } from "../../UIGlobals";
import {
  onDiscardStacksChanged,
  onHandsChanged,
  onHoleChanged,
  onPlayStacksChanged,
} from "./cardLayoutView";

jest.mock("../../UIGlobals", () => ({
  globals: {
    deck: [],
    lobby: { tableID: 123 },
    variant: { suits: ["Red", "Yellow", "Green", "Blue", "Purple"] },
    elements: {
      playerHands: [],
      discardStacks: new Map(),
      playStacks: new Map(),
    },
    layers: { card: { batchDraw: jest.fn() } },
    store: {
      getState: () => ({
        replay: {
          active: true,
          segment: 60,
          shared: null,
          databaseID: 1_576_519,
          hypothetical: {},
        },
        visibleState: { turn: { segment: 61 } },
      }),
    },
  },
}));
jest.mock("./cardsView", () => ({ updateCardVisuals: jest.fn() }));

function makeCard(order: number): HanabiCard {
  return {
    state: { order },
    isStackBase: false,
    parent: { moveUp: jest.fn(), moveDown: jest.fn() },
    animateToPlayerHand: jest.fn(),
  } as unknown as HanabiCard;
}

function setupCollection(orders: readonly number[]) {
  const children = orders.map((order) => {
    const card = makeCard(order);
    globals.deck[order] = card;
    return { card };
  });
  const container = { children: { toArray: () => children } };
  const elements = globals.elements as unknown as {
    playerHands: unknown[];
    discardStacks: Map<string, unknown>;
    playStacks: Map<string, unknown>;
  };
  const emptyContainer = { children: { toArray: () => [] } };
  elements.playerHands[0] = container;
  elements.playerHands[1] = emptyContainer;
  for (const [i, suit] of [
    "Red",
    "Yellow",
    "Green",
    "Blue",
    "Purple",
  ].entries()) {
    const stack = i === 0 ? container : emptyContainer;
    elements.discardStacks.set(suit, stack);
    elements.playStacks.set(suit, stack);
  }
  elements.playStacks.set("hole", container);
}

function getThrownError(callback: () => void): Error {
  try {
    callback();
  } catch (error) {
    if (error instanceof Error) {
      return error;
    }
    throw error;
  }
  throw new Error("Expected the card layout to reject a mismatch.");
}

describe("card layout synchronization diagnostics", () => {
  beforeEach(() => {
    globals.deck.length = 0;
    jest.clearAllMocks();
  });

  test.each([
    {
      collection: "hands[0]",
      run: () => {
        onHandsChanged([[0 as CardOrder], []]);
      },
    },
    {
      collection: "discardStacks[0]",
      run: () => {
        onDiscardStacksChanged([[0 as CardOrder], [], [], [], []]);
      },
    },
    {
      collection: "playStacks[0]",
      run: () => {
        onPlayStacksChanged([[0 as CardOrder], [], [], [], []], undefined);
      },
    },
    {
      collection: "hole",
      run: () => {
        onHoleChanged([0 as CardOrder], [1 as CardOrder]);
      },
    },
  ])(
    "identifies $collection when card counts diverge",
    ({ collection, run }) => {
      setupCollection([0, 0]);

      const error = getThrownError(run);

      expect(error.message).toContain(
        "The UI collection is out of sync with the state.",
      );
      expect(error.message).toContain(`Collection: ${collection}`);
      expect(error.message).toContain("Phase: reordering");
      expect(error.message).toContain("Expected orders: [0]");
      expect(error.message).toContain("Actual orders: [0,0]");
      expect(error.message).toContain("Game ID: 1576519");
      expect(error.message).toContain("Visible segment: 61");
      expect(error.message).toContain("Replay segment: 60");
      expect(error.message).toContain("Hypothetical: active");
    },
  );

  test("includes both orderings when the final positions differ", () => {
    setupCollection([1, 0]);

    const error = getThrownError(() => {
      onHandsChanged([[0 as CardOrder, 1 as CardOrder], []]);
    });

    expect(error.message).toContain("Collection: hands[0]");
    expect(error.message).toContain("Phase: final verification");
    expect(error.message).toContain("Expected orders: [0,1]");
    expect(error.message).toContain("Actual orders: [1,0]");
  });

  test("does not report an error for a synchronized hand", () => {
    setupCollection([0]);

    expect(() => {
      onHandsChanged([[0 as CardOrder], []]);
    }).not.toThrow();
  });
});
