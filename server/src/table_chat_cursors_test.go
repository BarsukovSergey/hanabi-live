package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTableChatMessageCursorAndUnread(t *testing.T) {
	table := &Table{LastReadMessageID: make(map[int]uint64)}
	first := &TableChatMessage{Msg: "first"}
	second := &TableChatMessage{Msg: "second"}
	table.AppendChat(first)
	table.AppendChat(second)
	if first.ID != 1 || second.ID != 2 || table.LastChatMessageID != 2 {
		t.Fatalf("table chat IDs should be monotonic: %+v, %+v", first, second)
	}
	if n := table.UnreadChatCount(42); n != 2 {
		t.Fatalf("Unread with no cursor = %d, want 2", n)
	}
	table.MarkChatRead(42, first.ID)
	if n := table.UnreadChatCount(42); n != 1 {
		t.Fatalf("Unread after first read = %d, want 1", n)
	}
	table.MarkChatRead(42, 0) // A delayed acknowledgement cannot regress the cursor.
	if table.LastReadMessageID[42] != 1 {
		t.Fatalf("Cursor regressed to %d", table.LastReadMessageID[42])
	}
	table.MarkChatRead(42, 1000) // A malicious/future cursor cannot skip unread chat.
	if table.LastReadMessageID[42] != 2 {
		t.Fatalf("Cursor advanced past latest: %d", table.LastReadMessageID[42])
	}
	table.Chat = table.Chat[1:] // Future pruning must not reuse earlier IDs.
	third := &TableChatMessage{Msg: "third"}
	table.AppendChat(third)
	if third.ID != 3 || table.UnreadChatCount(42) != 1 {
		t.Fatalf("Pruned chat IDs/unread not stable: %+v", table)
	}
}

func TestEmptyTableReadCursorSurvivesFirstRoutineNotice(t *testing.T) {
	table := &Table{LastReadMessageID: make(map[int]uint64)}
	table.MarkChatRead(25, table.LastChatMessageID)
	if _, exists := table.LastReadMessageID[25]; !exists {
		t.Fatal("Acknowledging an empty chat did not create a read cursor")
	}
	table.AppendChat(&TableChatMessage{Msg: "paused the game"})
	table.AdvanceCaughtUpReaders(0)
	if unread := table.UnreadChatCount(25); unread != 0 {
		t.Fatalf("First routine notice should already be read, got %d unread", unread)
	}
	if unread := table.UnreadChatCount(26); unread != 1 {
		t.Fatalf("Users without a read acknowledgement should see notice, got %d unread", unread)
	}
}

func TestTableChatSkipRoutineNoticeOnlyForCaughtUpReaders(t *testing.T) {
	table := &Table{LastReadMessageID: make(map[int]uint64)}
	table.AppendChat(&TableChatMessage{Msg: "first"})
	table.MarkChatRead(10, 1)
	table.MarkChatRead(11, 1)
	table.AppendChat(&TableChatMessage{Msg: "second"})
	table.MarkChatRead(10, 2)
	table.AppendChat(&TableChatMessage{Msg: "pause"})
	table.AdvanceCaughtUpReaders(2)
	if table.LastReadMessageID[10] != 3 || table.LastReadMessageID[11] != 1 {
		t.Fatalf("Pause notice should skip only caught-up cursors: %+v", table.LastReadMessageID)
	}
}

func TestTableChatRestoreLegacyReadOffsets(t *testing.T) {
	// This is the persisted JSON shape written by Go before PR9.
	oldJSON := `{"Chat":[
		{"UserID":1,"Username":"Ada","Msg":"first","Server":false},
		{"UserID":2,"Username":"Bob","Msg":"second","Server":false},
		{"UserID":1,"Username":"Ada","Msg":"third","Server":false}
	],"ChatRead":{"8":2,"9":100,"10":0}}`
	var table Table
	if err := json.Unmarshal([]byte(oldJSON), &table); err != nil {
		t.Fatal(err)
	}
	table.RestoreChatCursors()
	if table.LastChatMessageID != 3 || table.Chat[0].ID != 1 || table.Chat[1].ID != 2 || table.Chat[2].ID != 3 {
		t.Fatalf("Failed to assign stable IDs to legacy chat: %+v", table.Chat)
	}
	if table.LastReadMessageID[8] != 2 || table.LastReadMessageID[9] != 3 {
		t.Fatalf("Failed to convert read offsets to IDs: %+v", table.LastReadMessageID)
	}
	if table.UnreadChatCount(8) != 1 || table.UnreadChatCount(10) != 3 {
		t.Fatalf("Legacy unread semantics changed")
	}
	encoded, err := json.Marshal(&table)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"ChatRead"`) || !strings.Contains(string(encoded), `"LastReadMessageID"`) {
		t.Fatalf("Legacy ChatRead was not migrated in persisted JSON: %s", encoded)
	}
}

func TestTableChatRestoreRetainsSequenceAfterPruning(t *testing.T) {
	var table Table
	blob := `{"Chat":[{"ID":8,"Msg":"existing"}],"LastChatMessageID":10,"LastReadMessageID":{"3":8}}`
	if err := json.Unmarshal([]byte(blob), &table); err != nil {
		t.Fatal(err)
	}
	table.RestoreChatCursors()
	table.AppendChat(&TableChatMessage{Msg: "new"})
	if table.LastChatMessageID != 11 || table.Chat[1].ID != 11 || table.UnreadChatCount(3) != 1 {
		t.Fatalf("Failed to preserve cursor after restart: %+v", table)
	}
}
