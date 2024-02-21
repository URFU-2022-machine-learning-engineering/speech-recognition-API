package handlers

import (
	"github.com/rs/zerolog/log" // Import the zerolog package
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
		// Log the non-GET request method using zerolog
		log.Warn().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("Remote addr", r.RemoteAddr).
			Msg("Method Not Allowed")

		// Return a 405 Method Not Allowed response for non-GET requests
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := w.Write([]byte("Method Not Allowed"))
		if err != nil {
			// Log the error using zerolog
			log.Error().
				Err(err).
				Msg("Failed to write response")
			return
		}
	}
}
