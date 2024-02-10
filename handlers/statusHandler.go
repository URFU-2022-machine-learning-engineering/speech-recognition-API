package handlers

import (
	"fmt"
	"net/http"
	"sr-api/utils"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := utils.StartSpanFromRequest(r, "StatusHandler")
	defer span.End()
	// Check if the request method is GET
	if r.Method == "GET" {
		utils.RespondWithInfo(ctx, w, http.StatusOK)
		return
	} else {
		// Return a 405 Method Not Allowed response for non-GET requests
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := fmt.Fprintln(w, "Method Not Allowed")
		if err != nil {
			return
		}
	}
}
