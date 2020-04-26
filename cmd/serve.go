package cmd

import (
	"fmt"

	"github.com/jamesmichael/capture/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(serveCommand)

	serveCommand.Flags().String("addr", ":8080", "Address to serve on")
	viper.BindPFlag("addr", serveCommand.Flags().Lookup("addr"))

	serveCommand.Flags().String("template", "/var/lib/uk.jamesm.capture/index.html", "Path to template file")
	viper.BindPFlag("template", serveCommand.Flags().Lookup("template"))

	serveCommand.Flags().String("config", "/etc/uk.jamesm/capture.json", "Path to config file")
	viper.BindPFlag("config", serveCommand.Flags().Lookup("config"))
}

var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "Serve the webinterface",
	Run: func(cmd *cobra.Command, args []string) {

		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()

		sugar := logger.Sugar()

		server, err := server.New(
			server.WithAddress(viper.GetString("addr")),
			server.WithConfig(viper.GetString("config")),
			server.WithTemplate(viper.GetString("template")),
			server.WithLogger(sugar),
		)
		if err != nil {
			sugar.Fatalw("unable to instantiate server",
				"error", err,
			)
		}

		server.Serve()

		fmt.Println("Serving")
	},
}
