package cli

import (
	"context"
	"github.com/Conty111/AlfredoBot/internal/app"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

// NewServeCmd starts new application instance
func NewServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s"},
		Short:   "Start server",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().Msg("Starting")

			sigchan := make(chan os.Signal, 1)
			signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			application, err := app.InitializeApplication()

			if err != nil {
				log.Fatal().Err(err).Msg("can not initialize application")
			}

			cliMode := false
			application.Start(ctx, cliMode)

			log.Info().Msg("Started")
			<-sigchan

			log.Error().Err(application.Stop()).Msg("stop application")

			log.Info().Msg("Finished")
		},
	}
}
