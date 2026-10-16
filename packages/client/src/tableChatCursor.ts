// Table chat IDs are monotonic only within a table; they are not lobby DB IDs.
export function parseTableChatRoom(
  room: string | undefined,
): number | undefined {
  if (room === undefined || !/^table\d+$/u.test(room)) {
    return undefined;
  }
  const tableID = Number(room.slice("table".length));
  return Number.isSafeInteger(tableID) && tableID > 0 ? tableID : undefined;
}

export function getTableChatIDAttribute(
  room: string | undefined,
  tableChatID: number | undefined,
): string {
  const tableID = parseTableChatRoom(room);
  if (
    tableID === undefined
    || tableChatID === undefined
    || !Number.isSafeInteger(tableChatID)
    || tableChatID < 1
  ) {
    return "";
  }
  return ` data-table-chat-id="${tableChatID}" data-table-room="table${tableID}"`;
}

export function getTableChatIDSelector(
  room: string | undefined,
  tableChatID: number | undefined,
): string | undefined {
  const tableID = parseTableChatRoom(room);
  if (
    tableID === undefined
    || tableChatID === undefined
    || !Number.isSafeInteger(tableChatID)
    || tableChatID < 1
  ) {
    return undefined;
  }
  return `[data-table-chat-id="${tableChatID}"][data-table-room="table${tableID}"]`;
}

// A table's read cursor must never carry over to another table.
export class TableChatReadCursor {
  private tableID: number | undefined;
  private lastSeenID: number | undefined;

  public observe(
    room: string | undefined,
    tableChatID: number | undefined,
  ): void {
    const tableID = parseTableChatRoom(room);
    if (
      tableID === undefined
      || tableChatID === undefined
      || !Number.isSafeInteger(tableChatID)
      || tableChatID < 1
    ) {
      return;
    }
    if (this.tableID !== tableID) {
      this.tableID = tableID;
      this.lastSeenID = undefined;
    }
    if (this.lastSeenID === undefined || tableChatID > this.lastSeenID) {
      this.lastSeenID = tableChatID;
    }
  }

  public getLastSeenID(tableID: number): number | undefined {
    return this.tableID === tableID ? this.lastSeenID : undefined;
  }
}
