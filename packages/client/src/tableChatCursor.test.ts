import { describe, expect, test } from "@jest/globals";
import {
  getTableChatIDSelector,
  parseTableChatRoom,
  TableChatReadCursor,
} from "./tableChatCursor";

describe("table-local chat message identity", () => {
  test("parses valid table rooms, including leading zeros", () => {
    expect(parseTableChatRoom("table00042")).toBe(42);
    expect(parseTableChatRoom("table42")).toBe(42);
    expect(parseTableChatRoom("lobby")).toBeUndefined();
    expect(parseTableChatRoom("table0")).toBeUndefined();
    expect(parseTableChatRoom("table-1")).toBeUndefined();
    expect(parseTableChatRoom("table1x")).toBeUndefined();
  });

  test("the dedupe selector normalizes leading-zero room names", () => {
    expect(getTableChatIDSelector("table00042", 2)).toBe(
      '[data-table-chat-id="2"][data-table-room="table42"]',
    );
    expect(getTableChatIDSelector("table43", 2)).not.toBe(
      getTableChatIDSelector("table42", 2),
    );
    expect(getTableChatIDSelector("lobby", 2)).toBeUndefined();
  });

  test("only the highest actually observed message ID is acknowledged", () => {
    const state = new TableChatReadCursor();
    state.observe("table3", 10);
    state.observe("table3", 8);
    expect(state.getLastSeenID(3)).toBe(10);
    expect(state.getLastSeenID(4)).toBeUndefined();
    state.observe("table3", 11);
    expect(state.getLastSeenID(3)).toBe(11);
  });

  test("switching tables discards the old table's cursor", () => {
    const state = new TableChatReadCursor();
    state.observe("table1", 20);
    state.observe("table2", 1);
    expect(state.getLastSeenID(1)).toBeUndefined();
    expect(state.getLastSeenID(2)).toBe(1);
  });

  test("old servers without table IDs do not change the cursor", () => {
    const state = new TableChatReadCursor();
    state.observe("table7", undefined);
    expect(state.getLastSeenID(7)).toBeUndefined();
    state.observe("table7", 4);
    state.observe("table7", undefined);
    expect(state.getLastSeenID(7)).toBe(4);
  });
});
