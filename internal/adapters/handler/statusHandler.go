package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"sr-api/internal/core/ports/telemetry"
)

func StatusHandler(c *gin.Context) {
	_, span := telemetry.StartSpanFromGinContext(c, "StatusHandler")
	defer span.End()
	spanID := telemetry.GetSpanId(span)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
	log.Info().Str("span_id", spanID).Str("method", c.Request.Method).Msg("Status endpoint hit")
}
