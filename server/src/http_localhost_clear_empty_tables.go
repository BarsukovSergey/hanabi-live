package main

import (
	"net/http"
	"strconv"

	"github.com/Hanabi-Live/hanabi-live/logger"
	"github.com/gin-gonic/gin"
)

// If calls to the database fail for whatever reason,
// it is possible for tables to be created with no people in them
// So we allow an administrator to clear them manually
func httpLocalhostClearEmptyTables(c *gin.Context) {
	for _, t := range tables.GetList(true) {
		t.Lock(c)
		if !t.Deleted {
			if !t.Running && len(t.Players) == 0 {
				// A table that has not started yet (e.g. pregame)
				tables.Lock(c)
				deleteTable(t)
				tables.Unlock(c)
				logger.Info("Successfully cleared pregame table #" + strconv.FormatUint(t.ID, 10) + ".")
			} else if t.Replay && len(t.ActiveSpectators()) == 0 {
				// A replay or shared replay
				tables.Lock(c)
				deleteTable(t)
				tables.Unlock(c)
				logger.Info("Successfully cleared replay table #" + strconv.FormatUint(t.ID, 10) + ".")
			}
			// (do not do anything for ongoing games)
		}
		t.Unlock(c)
	}

	c.String(http.StatusOK, "success\n")
}
