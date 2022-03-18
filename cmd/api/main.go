package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"go-server/cmd"
	"go-server/pkg/routing"
	"net/http"
	"os"
)

func main() {
	pc, sc := cmd.ParseArgs()
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()

	r := routing.GetRouter(pc, logger)

	logger.Info().Msgf("starting api: %s:%d", sc.ListenHost, sc.ListenPort)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", sc.ListenHost, sc.ListenPort), r)
	if err != nil {
		logger.Err(err).Msgf("failed to listen port: %d", sc.ListenPort)
		return
	}
}
