package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/Conty111/AlfredoBot/internal/app"
	"github.com/Conty111/AlfredoBot/internal/configs"
)

// NewServeCmd starts new application instance
func NewServeCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s"},
		Short:   "Start server",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().Msg("Starting")

			sigchan := make(chan os.Signal, 1)
			signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg, err := configs.LoadConfig(configPath)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load configuration")
			}
			application, err := app.InitializeApplication(cfg)
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

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default searches for config.yaml|json)")

	return cmd
}
