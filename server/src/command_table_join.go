package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Hanabi-Live/hanabi-live/logger"
	"github.com/alexedwards/argon2id"
)

// commandTableJoin is sent when the user clicks on the "Join" button in the lobby
//
// Example data:
//
//	{
//	  tableID: 15103,
//	}
func commandTableJoin(ctx context.Context, s *Session, d *CommandData) {
	t, exists := getTableAndLock(ctx, s, d.TableID, !d.NoTableLock, !d.NoTablesLock)
	if !exists {
		return
	}
	tableLocked := !d.NoTableLock
	if !d.NoTableLock {
		defer func() {
			if tableLocked {
				t.Unlock(ctx)
			}
		}()
	}

	if !validateTableJoinState(s, t) {
		return
	}

	if d.PregameStats == nil {
		if d.NoTableLock {
			logger.Error("commandTableJoin was called without pre-fetched stats while the caller held the table lock.")
			s.Error(DefaultErrorMsg)
			return
		}

		passwordHash := t.PasswordHash
		variantName := t.Options.VariantName
		t.Unlock(ctx)
		tableLocked = false

		if !validateTableJoinPassword(s, d, passwordHash) {
			return
		}

		numGames, err := models.Games.GetUserNumGames(s.UserID, false)
		if err != nil {
			logger.Error("Failed to pre-fetch the number of non-speedrun games for player " +
				"\"" + s.Username + "\": " + err.Error())
			s.Error("Something went wrong when getting your stats. Please contact an administrator.")
			return
		}

		for {
			variant := variants[variantName]
			variantStats, err := models.UserStats.Get(s.UserID, variant.ID)
			if err != nil {
				logger.Error("Failed to pre-fetch variant stats for player \"" + s.Username +
					"\" for variant " + strconv.Itoa(variant.ID) + ": " + err.Error())
				s.Error("Something went wrong when getting your stats. Please contact an administrator.")
				return
			}

			reloadedTable, tableExists := getTableAndLock(ctx, s, d.TableID, true, !d.NoTablesLock)
			if !tableExists {
				return
			}
			t = reloadedTable
			tableLocked = true

			if !validateTableJoinState(s, t) {
				return
			}
			if t.Options.VariantName == variantName {
				d.PregameStats = &PregameStats{NumGames: numGames, Variant: variantStats}
				break
			}

			variantName = t.Options.VariantName
			t.Unlock(ctx)
			tableLocked = false
		}
	} else if !validateTableJoinPassword(s, d, t.PasswordHash) {
		return
	}

	tableJoin(ctx, s, d, t)
}

func validateTableJoinState(s *Session, t *Table) bool {
	// Validate that the player is not already joined to this table
	playerIndex := t.GetPlayerIndexFromID(s.UserID)
	if playerIndex != -1 {
		s.Warning("You have already joined this table.")
		return false
	}

	// Validate that this table does not already have the maximum number of players
	if len(t.Players) >= t.MaxPlayers {
		s.Warning("That table is already full.")
		return false
	}

	// Validate that the game is not started yet
	if t.Running {
		s.Warning("That game has already started, so you cannot join it.")
		return false
	}

	// Validate that it is not a replay
	if t.Replay {
		s.Warning("You can not join a replay.")
		return false
	}

	// Validate that they have not been previously kicked from this game
	if _, ok := t.KickedPlayers[s.UserID]; ok {
		s.Warning("You cannot join a game that you have been kicked from.")
		return false
	}

	return true
}

func validateTableJoinPassword(s *Session, d *CommandData, passwordHash string) bool {
	if passwordHash == "" || d.BypassPassword {
		return true
	}

	match, err := argon2id.ComparePasswordAndHash(d.Password, passwordHash)
	if err != nil {
		logger.Error("Failed to compare the submitted password to the Argon2 hash: " + err.Error())
		s.Error(DefaultErrorMsg)
		return false
	}
	if !match {
		s.Warning("That is not the correct password for this game.")
		return false
	}

	return true
}

func tableJoin(ctx context.Context, s *Session, d *CommandData, t *Table) {
	// Since this is a function that changes a user's relationship to tables,
	// we must acquires the tables lock to prevent race conditions
	if !d.NoTablesLock {
		tables.Lock(ctx)
		defer tables.Unlock(ctx)
	}

	// Validate that the player is not joined to any table
	// (this cannot be in the "commandTableJoin()" function because we need the tables lock)
	// (only bots have the ability to join more than one table)
	if !strings.HasPrefix(s.Username, "Bot-") {
		if len(tables.GetTablesUserPlaying(s.UserID)) > 0 {
			s.Warning("You cannot join more than one table at a time. " +
				"Terminate your other game before joining a new one.")
			return
		}
	}

	logger.Info(t.GetName() + "User \"" + s.Username + "\" joined. " +
		"(There are now " + strconv.Itoa(len(t.Players)+1) + " players.)")

	p := &Player{
		UserID:     s.UserID,
		Name:       s.Username,
		Session:    s,
		Present:    true,
		Stats:      d.PregameStats,
		Typing:     false,
		LastTyped:  time.Time{},
		VoteToKill: false,
	}

	t.Players = append(t.Players, p)
	tables.AddPlaying(s.UserID, t.ID) // Keep track of user to table relationships

	// Sanity check, that we are not actively spectating this table
	// Might occur if a command to joinTable happens after a tableSpectate, without tableUnattend in between
	if t.IsActivelySpectating(s.UserID) {
		tables.DeleteSpectating(s.UserID, t.ID) // Update user to table relationships
	}

	notifyAllTable(t)
	t.NotifyPlayerChange()

	// Set their status
	s.SetStatus(StatusPregame)
	s.SetTableID(t.ID)
	notifyAllUser(s)

	// Let the client know they successfully joined the table
	s.NotifyTableJoined(t)

	// Send them the chat history for this game
	chatSendPastFromTable(s, t)
	t.ChatRead[p.UserID] = len(t.Chat)

	// Announce the new player (unless this is the first one)
	if len(t.Players) > 1 {
		msg := s.Username + " joined the game."
		chatServerSend(ctx, msg, t.GetRoomName(), true)
	}

	// Send them the list of spectators
	s.NotifySpectators(t)

	// Send them messages for people typing, if any
	for _, p := range t.Players {
		if p.Typing {
			s.NotifyChatTyping(t, p.Name, p.Typing)
		}
	}

	// If there is an automatic start countdown, cancel it
	if !t.DatetimePlannedStart.IsZero() {
		t.DatetimePlannedStart = time.Time{} // Assign a zero value
		msg := "Automatic game start has been canceled."
		chatServerSend(ctx, msg, t.GetRoomName(), true)
	}

	// If the user previously requested it, automatically start the game
	if t.AutomaticStart == len(t.Players) {
		// Check to see if the owner is present
		for _, p2 := range t.Players {
			if p2.UserID == t.OwnerID {
				if !p2.Present {
					msg := "Aborting automatic game start since the table creator is away."
					chatServerSend(ctx, msg, t.GetRoomName(), true)
					return
				}

				commandTableStart(ctx, p2.Session, &CommandData{ // nolint: exhaustivestruct
					TableID:      t.ID,
					NoTableLock:  true,
					NoTablesLock: true,
				})
				return
			}
		}

		logger.Error("Failed to find the owner of the game when attempting to automatically start it.")
		return
	}

	// Update the "DatetimeLastJoined" field, but make a copy first
	datetimeLastJoined := t.DatetimeLastJoined
	t.DatetimeLastJoined = time.Now()

	// Play a notification sound if it has been more than 15 seconds since the last person joined
	if time.Since(datetimeLastJoined) > time.Second*15 {
		for _, p2 := range t.Players {
			// Skip sending a message to the player that just joined
			if p2.UserID != p.UserID {
				p2.Session.NotifySoundLobby("someone-joined")
			}
		}
	}
}
