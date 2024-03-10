package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"sr-api/helpers"
)

func StatusHandler(c *gin.Context) {
	_, span := helpers.StartSpanFromGinContext(c, "StatusHandler")
	defer span.End()

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
	log.Info().Str("method", c.Request.Method).Msg("Status endpoint hit")
}
