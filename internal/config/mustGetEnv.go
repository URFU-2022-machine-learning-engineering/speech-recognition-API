package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatal().Err(fmt.Errorf(key + " environment variable is not set")).Msg(key + " environment variable is not set")
		return ""
	}
	log.Debug().Msgf("Got a non empty variable: '%s'", key)
	return value
}
