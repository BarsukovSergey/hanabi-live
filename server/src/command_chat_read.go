package main

import (
	"context"
)

// commandChatRead is sent when the user opens the in-game chat or
// when they receive a chat message when the in-game chat is already open
//
// Example data:
//
//	{
//	  tableID: 5,
//	}
func commandChatRead(ctx context.Context, s *Session, d *CommandData) {
	// Unlike other command handlers, we do not want to show a warning to the user if the table does
	// not exist, so we pass "nil" instead of "s" to the "getTableAndLock()" function
	// This is because in some cases, network latency will cause the "read" message to get to
	// the server after the respective table has already been deleted
	t, exists := getTableAndLock(ctx, nil, d.TableID, !d.NoTableLock, !d.NoTablesLock)
	if !exists {
		return
	}
	if !d.NoTableLock {
		defer t.Unlock(ctx)
	}

	// Validate that they are in the game or are a spectator
	if !t.IsPlayerOrSpectating(s.UserID) {
		// Return without an error message if they are not playing or spectating at the table
		// (to account for lag)
		return
	}
	if !t.IsActivelySpectating(s.UserID) && t.Replay {
		// Return without an error message if they are not spectating at the replay
		// (to account for lag)
		return
	}

	// New clients acknowledge the last actually received table message,
	// so an in-flight message cannot accidentally be marked read.
	// Older browser tabs omit the cursor; preserve their mark-all behavior.
	if d.LastReadMessageID > 0 {
		t.MarkChatRead(s.UserID, d.LastReadMessageID)
	} else {
		t.MarkChatRead(s.UserID, t.LastChatMessageID)
	}
}
