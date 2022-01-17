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
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	pc, sc := cmd.ParseArgs()

	r := routing.GetRouter(pc, logger)

	logger.Info().Msgf("starting api: [::]:%d", sc.ListenPort)

	err := http.ListenAndServe(fmt.Sprintf(":%d", sc.ListenPort), r)
	if err != nil {
		logger.Err(err).Msgf("failed to listen port: %d", sc.ListenPort)
		return
	}
}
